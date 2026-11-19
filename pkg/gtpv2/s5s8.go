package gtpv2

import (
	"fmt"
	"net"
	"time"

	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/ie"
	"github.com/wmnsk/go-gtp/gtpv2/message"
)

// localHost returns the host portion of a net.Addr that carries a host:port
// (or [host]:port for IPv6). Falls back to the raw string when SplitHostPort
// cannot parse it, so ergonomics for callers passing bare IPs are preserved.
func localHost(a net.Addr) string {
	s := a.String()
	if host, _, err := net.SplitHostPort(s); err == nil {
		return host
	}
	return s
}

type S5S8SgwParams struct {
	Imsi        string
	Msisdn      string
	Mei         string
	Mcc         string
	Mnc         string
	Tac         uint16
	Ecgi        string
	Rat         string
	Ftei        string
	Apn         string
	Eci         uint32
	Epsbearerid uint8
	PdnType     uint8 // 1:v4, 2:v6, 3:v4v6, 4:non ip
	UplaneIE    TEIDParams
	CplaneSgwIE TEIDParams
	CplanePgwIE TEIDParams
	Ambrul      uint32
	Ambrdl      uint32

	// IfTypeCplane overrides the FTEID interface type used for the
	// local-side C-plane bearer. Pointers so that an explicit 0
	// (S1U_ENODEB_GTPU) is distinguishable from "not set". When unset, an
	// auto-allocated TEID keeps the connection's interface type and an
	// explicit TEID uses IFTypeS5S8SGWGTPC (6).
	IfTypeCplane *uint8
	// IfTypeUplane overrides the FTEID interface type used for the
	// local-side U-plane bearer. Defaults to IFTypeS5S8SGWGTPU (4).
	IfTypeUplane *uint8
	// IfTypeCplanePeer overrides the FTEID interface type expected on the
	// remote side (used by Delete / Modify session lookups). Defaults to
	// IFTypeS5S8PGWGTPC (7).
	IfTypeCplanePeer *uint8
}

func ifTypeOr(v *uint8, def uint8) uint8 {
	if v == nil {
		return def
	}
	return *v
}

func (p S5S8SgwParams) cplaneIfType() uint8 {
	return ifTypeOr(p.IfTypeCplane, gtpv2.IFTypeS5S8SGWGTPC)
}

func (p S5S8SgwParams) uplaneIfType() uint8 {
	return ifTypeOr(p.IfTypeUplane, gtpv2.IFTypeS5S8SGWGTPU)
}

func (p S5S8SgwParams) cplanePeerIfType() uint8 {
	return ifTypeOr(p.IfTypeCplanePeer, gtpv2.IFTypeS5S8PGWGTPC)
}

type TEIDParams struct {
	Teid uint32
	IP   string
	IP6  string
}

func (c *K6GTPv2Client) SendCreateSessionRequestS5S8(daddr string, options S5S8SgwParams) (*gtpv2.Session, uint32, error) {
	sess, seq, err := c.sendCreateSessionS5S8(daddr, options)
	if err != nil {
		c.rm.metrics.pushSendError(c.iterCtx(), c.vu.State(), msgTypeCreateSession)
		return nil, seq, err
	}
	c.rm.metrics.pushRequest(c.iterCtx(), c.vu.State(), msgTypeCreateSession)
	return sess, seq, nil
}

func (c *K6GTPv2Client) sendCreateSessionS5S8(daddr string, options S5S8SgwParams) (*gtpv2.Session, uint32, error) {
	if err := c.connReady(); err != nil {
		return nil, 0, err
	}
	d, err := net.ResolveUDPAddr("udp", daddr)
	if err != nil {
		return nil, 0, fmt.Errorf("resolve destination UDP addr %q: %w", daddr, err)
	}

	localIP := localHost(c.Conn.LocalAddr())

	// todo v6
	cteidIE, err := c.senderCplaneFTEID(options, localIP)
	if err != nil {
		return nil, 0, err
	}
	uteidIE := ie.NewFullyQualifiedTEID(options.uplaneIfType(), options.UplaneIE.Teid, localIP, "").WithInstance(2) // dummy uplane teid
	if uteidIE == nil {
		return nil, 0, fmt.Errorf("uteidIE is null, unexpected error")
	}
	ies := c.genS5S8SessionIE(options, cteidIE, uteidIE, localIP)
	sess, err := c.Conn.ParseCreateSession(d, ies...)
	if err != nil {
		return nil, 0, err
	}
	sess.AddTEID(uteidIE.MustInterfaceType(), uteidIE.MustTEID())
	// Register before dispatch so the reply cannot beat the registration.
	c.Conn.RegisterSession(cteidIE.MustTEID(), sess)
	// RegisterSession records the TEID under the connection's interface
	// type; also record it under the advertised one for later lookups.
	sess.AddTEID(cteidIE.MustInterfaceType(), cteidIE.MustTEID())

	seq, err := c.Conn.SendMessageTo(message.NewCreateSessionRequest(0, 0, ies...), d)
	if err != nil {
		c.Conn.RemoveSession(sess)
		return nil, seq, fmt.Errorf("send Create Session Request: %w", err)
	}
	return sess, seq, nil
}

// senderCplaneFTEID builds the local C-plane F-TEID. A zero TEID is
// allocated by go-gtp (unique within the Conn) and then re-typed when
// IfTypeCplane is set.
func (c *K6GTPv2Client) senderCplaneFTEID(options S5S8SgwParams, localIP string) (*ie.IE, error) {
	if options.CplaneSgwIE.Teid != 0 {
		return ie.NewFullyQualifiedTEID(options.cplaneIfType(), options.CplaneSgwIE.Teid, localIP, ""), nil
	}
	auto := c.Conn.NewSenderFTEID(localIP, "")
	if auto == nil {
		return nil, fmt.Errorf("gtpv2: could not allocate a local C-plane TEID")
	}
	if options.IfTypeCplane == nil {
		return auto, nil
	}
	return ie.NewFullyQualifiedTEID(*options.IfTypeCplane, auto.MustTEID(), localIP, ""), nil
}

func (c *K6GTPv2Client) SendDeleteSessionRequestS5S8(daddr string, options S5S8SgwParams) (uint32, error) {
	seq, err := c.sendDeleteSessionS5S8(daddr, options)
	if err != nil {
		c.rm.metrics.pushSendError(c.iterCtx(), c.vu.State(), msgTypeDeleteSession)
		return seq, err
	}
	c.rm.metrics.pushRequest(c.iterCtx(), c.vu.State(), msgTypeDeleteSession)
	return seq, nil
}

func (c *K6GTPv2Client) sendDeleteSessionS5S8(daddr string, options S5S8SgwParams) (uint32, error) {
	if err := c.connReady(); err != nil {
		return 0, err
	}
	var (
		s5Session *gtpv2.Session
		err       error
	)

	// gen s5 session
	if daddr != "" {
		// case only when delete session request
		s5Session, err = c.registerDummyS5S8Session(daddr, options)
		if err != nil {
			return 0, err
		}
	} else {
		// case when delete session has already been generated with create session
		s5Session, err = c.Conn.GetSessionByIMSI(options.Imsi)
		if err != nil {
			return 0, err
		}
		// teid override
		if options.CplanePgwIE.Teid == 0 {
			options.CplanePgwIE.Teid, err = s5Session.GetTEID(options.cplanePeerIfType())
			if err != nil {
				return 0, err
			}
		}
	}
	return c.Conn.DeleteSession(
		options.CplanePgwIE.Teid,
		s5Session,
		ie.NewEPSBearerID(options.Epsbearerid),
	)
}

func (c *K6GTPv2Client) registerDummyS5S8Session(daddr string, options S5S8SgwParams) (*gtpv2.Session, error) {
	d, err := net.ResolveUDPAddr("udp", daddr)
	if err != nil {
		return nil, fmt.Errorf("resolve destination UDP addr %q: %w", daddr, err)
	}
	localIP := localHost(c.Conn.LocalAddr())
	cteidIE := ie.NewFullyQualifiedTEID(options.cplaneIfType(), options.CplaneSgwIE.Teid, localIP, "")
	uteidIE := ie.NewFullyQualifiedTEID(options.uplaneIfType(), options.UplaneIE.Teid, localIP, "").WithInstance(2) // dummy uplane teid
	s5Session, err := c.Conn.ParseCreateSession(d, c.genS5S8SessionIE(options, cteidIE, uteidIE, localIP)...)
	if err != nil {
		return nil, err
	}

	s5Session.AddTEID(uteidIE.MustInterfaceType(), uteidIE.MustTEID())
	c.Conn.RegisterSession(cteidIE.MustTEID(), s5Session)
	return s5Session, nil
}

func (c *K6GTPv2Client) SendModifyBearerRequestS5S8(daddr string, options S5S8SgwParams) (uint32, error) {
	seq, err := c.sendModifyBearerS5S8(daddr, options)
	if err != nil {
		c.rm.metrics.pushSendError(c.iterCtx(), c.vu.State(), msgTypeModifyBearer)
		return seq, err
	}
	c.rm.metrics.pushRequest(c.iterCtx(), c.vu.State(), msgTypeModifyBearer)
	return seq, nil
}

func (c *K6GTPv2Client) sendModifyBearerS5S8(daddr string, options S5S8SgwParams) (uint32, error) {
	if err := c.connReady(); err != nil {
		return 0, err
	}
	var (
		s5Session *gtpv2.Session
		err       error
	)

	// gen s5 session
	if daddr != "" {
		// case only when delete session request
		s5Session, err = c.registerDummyS5S8Session(daddr, options)
		if err != nil {
			return 0, err
		}
	} else {
		// case when delete session has already been generated with create session
		s5Session, err = c.Conn.GetSessionByIMSI(options.Imsi)
		if err != nil {
			return 0, err
		}
		// teid override
		if options.CplanePgwIE.Teid == 0 {
			options.CplanePgwIE.Teid, err = s5Session.GetTEID(options.cplanePeerIfType())
			if err != nil {
				return 0, err
			}
		}
		if options.CplaneSgwIE.Teid == 0 {
			options.CplaneSgwIE.Teid, err = s5Session.GetTEID(options.cplaneIfType())
			if err != nil {
				return 0, err
			}
		}
		if options.UplaneIE.Teid == 0 {
			options.UplaneIE.Teid, err = s5Session.GetTEID(options.uplaneIfType())
			if err != nil {
				return 0, err
			}
		}
	}

	return c.Conn.ModifyBearer(
		options.CplanePgwIE.Teid,
		s5Session,
		ie.NewUserLocationInformationStruct(
			nil, nil, nil,
			ie.NewTAI(options.Mcc, options.Mnc, options.Tac),
			ie.NewECGI(options.Mcc, options.Mnc, options.Eci),
			nil, nil, nil,
		),
		ie.NewServingNetwork(options.Mcc, options.Mnc),
		ie.NewRATType(gtpv2.RATTypeEUTRAN),

		// todo Support multiple cplane teids
		ie.NewFullyQualifiedTEID(
			options.cplaneIfType(),
			options.CplaneSgwIE.Teid,
			options.CplaneSgwIE.IP,
			options.CplaneSgwIE.IP6,
		),
		ie.NewAggregateMaximumBitRate(options.Ambrul, options.Ambrdl),
		ie.NewMobileEquipmentIdentity(options.Mei),
		ie.NewUETimeZone(9, 0),

		ie.NewBearerContext(
			ie.NewEPSBearerID(options.Epsbearerid),
			ie.NewFullyQualifiedTEID(
				options.uplaneIfType(),
				options.UplaneIE.Teid,
				options.UplaneIE.IP,
				options.UplaneIE.IP6,
			).WithInstance(1),
		),
		ie.NewRecovery(0),
	)
}

func (c *K6GTPv2Client) genS5S8SessionIE(options S5S8SgwParams, cteidIE, uteidIE *ie.IE, localIP string) []*ie.IE {
	ielist := []*ie.IE{
		ie.NewIMSI(options.Imsi),
		ie.NewMSISDN(options.Msisdn),
		ie.NewMobileEquipmentIdentity(options.Mei),
		ie.NewAccessPointName(options.Apn),
		ie.NewServingNetwork(options.Mcc, options.Mnc),
		ie.NewRATType(gtpv2.RATTypeEUTRAN),
		cteidIE,
		ie.NewSelectionMode(gtpv2.SelectionModeMSorNetworkProvidedAPNSubscribedVerified),
		ie.NewPDNType(options.PdnType),
		ie.NewPDNAddressAllocation("0.0.0.0"),
		ie.NewAPNRestriction(gtpv2.APNRestrictionPublic2),
		ie.NewAggregateMaximumBitRate(options.Ambrul, options.Ambrdl),
		ie.NewUserLocationInformationStruct(
			nil, nil, nil,
			ie.NewTAI(options.Mcc, options.Mnc, options.Tac),
			ie.NewECGI(options.Mcc, options.Mnc, options.Eci),
			nil, nil, nil,
		),
		ie.NewBearerContext(
			ie.NewEPSBearerID(options.Epsbearerid),
			uteidIE,
			ie.NewBearerQoS(1, 2, 1, 0xff, 0, 0, 0, 0),
		),
		ie.NewFullyQualifiedCSID(localIP, 1).WithInstance(1),
	}
	return ielist
}

func (c *K6GTPv2Client) CheckSendCreateSessionRequestS5S8(daddr string, options S5S8SgwParams) (EnumIFCause, error) {
	start := time.Now()
	_, seq, err := c.SendCreateSessionRequestS5S8(daddr, options)
	if err != nil {
		c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeCreateSession, time.Since(start), true)
		return 0, err
	}
	res, err := c.CheckRecvCreateSessionResponse(seq, options.Imsi)
	c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeCreateSession, time.Since(start), err != nil)
	return res, err
}

func (c *K6GTPv2Client) CheckSendDeleteSessionRequestS5S8(daddr string, options S5S8SgwParams) (EnumIFCause, error) {
	start := time.Now()
	seq, err := c.SendDeleteSessionRequestS5S8(daddr, options)
	if err != nil {
		c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeDeleteSession, time.Since(start), true)
		return 0, err
	}
	res, err := c.CheckRecvDeleteSessionResponse(seq)
	c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeDeleteSession, time.Since(start), err != nil)
	return res, err
}

func (c *K6GTPv2Client) CheckSendModifyBearerRequestS5S8(daddr string, options S5S8SgwParams) (EnumIFCause, error) {
	start := time.Now()
	seq, err := c.SendModifyBearerRequestS5S8(daddr, options)
	if err != nil {
		c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeModifyBearer, time.Since(start), true)
		return 0, err
	}
	res, err := c.CheckRecvModifyBearerResponse(seq)
	c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeModifyBearer, time.Since(start), err != nil)
	return res, err
}

// TryCreateSessionS5S8 sends a Create Session Request over S5/S8 and waits for
// the paired response. Unlike CheckSendCreateSessionRequestS5S8, the returned
// SendResult exposes the Cause value, the elapsed round-trip time, and a
// Timeout flag so scripts can build flexible sequences without reaching into
// CheckRecv* helpers directly.
func (c *K6GTPv2Client) TryCreateSessionS5S8(daddr string, options S5S8SgwParams) *SendResult {
	start := time.Now()
	_, seq, err := c.SendCreateSessionRequestS5S8(daddr, options)
	if err != nil {
		c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeCreateSession, time.Since(start), true)
		return newSendError(time.Since(start), err)
	}
	cause, recvErr := c.CheckRecvCreateSessionResponse(seq, options.Imsi)
	elapsed := time.Since(start)
	c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeCreateSession, elapsed, recvErr != nil)
	if recvErr != nil {
		r := newSendError(elapsed, recvErr)
		r.Sequence = seq
		return r
	}
	return newSendOK(seq, uint8(cause), elapsed)
}

// TryDeleteSessionS5S8 mirrors TryCreateSessionS5S8 for Delete Session.
func (c *K6GTPv2Client) TryDeleteSessionS5S8(daddr string, options S5S8SgwParams) *SendResult {
	start := time.Now()
	seq, err := c.SendDeleteSessionRequestS5S8(daddr, options)
	if err != nil {
		c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeDeleteSession, time.Since(start), true)
		return newSendError(time.Since(start), err)
	}
	cause, recvErr := c.CheckRecvDeleteSessionResponse(seq)
	elapsed := time.Since(start)
	c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeDeleteSession, elapsed, recvErr != nil)
	if recvErr != nil {
		r := newSendError(elapsed, recvErr)
		r.Sequence = seq
		return r
	}
	return newSendOK(seq, uint8(cause), elapsed)
}

// TryModifyBearerS5S8 mirrors TryCreateSessionS5S8 for Modify Bearer.
func (c *K6GTPv2Client) TryModifyBearerS5S8(daddr string, options S5S8SgwParams) *SendResult {
	start := time.Now()
	seq, err := c.SendModifyBearerRequestS5S8(daddr, options)
	if err != nil {
		c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeModifyBearer, time.Since(start), true)
		return newSendError(time.Since(start), err)
	}
	cause, recvErr := c.CheckRecvModifyBearerResponse(seq)
	elapsed := time.Since(start)
	c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeModifyBearer, elapsed, recvErr != nil)
	if recvErr != nil {
		r := newSendError(elapsed, recvErr)
		r.Sequence = seq
		return r
	}
	return newSendOK(seq, uint8(cause), elapsed)
}

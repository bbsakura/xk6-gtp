package gtpv2

import (
	"fmt"
	"net"
	"strings"

	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/ie"
)

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
	UplaneIE    TEIDParams
	CplaneSgwIE TEIDParams
	CplanePgwIE TEIDParams
	Ambrul      uint32
	Ambrdl      uint32
}

type TEIDParams struct {
	Teid uint32
	IP   string
	IP6  string
}

func (c *K6GTPv2Client) SendCreateSessionRequestS5S8(daddr string, options S5S8SgwParams) (*gtpv2.Session, uint32, error) {
	d, err := net.ResolveUDPAddr("udp", daddr)
	if err != nil {
		return nil, 0, fmt.Errorf("resolve udp error")
	}

	localIP := strings.Split(c.Conn.LocalAddr().String(), ":")[0]

	// todo v6
	var cteidIE *ie.IE
	if options.CplaneSgwIE.Teid == 0 {
		cteidIE = c.Conn.NewSenderFTEID(localIP, "")
	} else {
		cteidIE = ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8SGWGTPC, options.CplaneSgwIE.Teid, localIP, "")
	}
	uteidIE := ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8SGWGTPU, options.UplaneIE.Teid, localIP, "").WithInstance(2) // dummy uplane teid

	sess, seq, err := c.Conn.CreateSession(d,
		c.genS5S8SessionIE(options, cteidIE, uteidIE, localIP)...,
	)
	sess.AddTEID(uteidIE.MustInterfaceType(), uteidIE.MustTEID())
	c.Conn.RegisterSession(cteidIE.MustTEID(), sess)
	return sess, seq, err
}

func (c *K6GTPv2Client) SendDeleteSessionRequestS5S8(daddr string, options S5S8SgwParams) (uint32, error) {
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
			options.CplanePgwIE.Teid, err = s5Session.GetTEID(gtpv2.IFTypeS5S8PGWGTPC)
			if err != nil {
				return 0, err
			}
		}
	}
	seq, err := c.Conn.DeleteSession(
		options.CplanePgwIE.Teid,
		s5Session,
		ie.NewEPSBearerID(options.Epsbearerid),
	)
	return seq, err
}

func (c *K6GTPv2Client) registerDummyS5S8Session(daddr string, options S5S8SgwParams) (*gtpv2.Session, error) {
	d, err := net.ResolveUDPAddr("udp", daddr)
	if err != nil {
		return nil, fmt.Errorf("resolve udp error")
	}
	localIP := strings.Split(c.Conn.LocalAddr().String(), ":")[0]
	cteidIE := ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8SGWGTPC, options.CplaneSgwIE.Teid, localIP, "")
	uteidIE := ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8SGWGTPU, options.UplaneIE.Teid, localIP, "").WithInstance(2) // dummy uplane teid
	s5Session, err := c.Conn.ParseCreateSession(d, c.genS5S8SessionIE(options, cteidIE, uteidIE, localIP)...)
	if err != nil {
		return nil, err
	}

	s5Session.AddTEID(uteidIE.MustInterfaceType(), uteidIE.MustTEID())
	c.Conn.RegisterSession(cteidIE.MustTEID(), s5Session)
	return s5Session, nil
}

func (c *K6GTPv2Client) SendModifyBearerRequestS5S8(daddr string, options S5S8SgwParams) (uint32, error) {
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
			options.CplanePgwIE.Teid, err = s5Session.GetTEID(gtpv2.IFTypeS5S8PGWGTPC)
			if err != nil {
				return 0, err
			}
		}
		if options.CplaneSgwIE.Teid == 0 {
			options.CplaneSgwIE.Teid, err = s5Session.GetTEID(gtpv2.IFTypeS5S8SGWGTPC)
			if err != nil {
				return 0, err
			}
		}
		if options.UplaneIE.Teid == 0 {
			options.UplaneIE.Teid, err = s5Session.GetTEID(gtpv2.IFTypeS5S8SGWGTPU)
			if err != nil {
				return 0, err
			}
		}
	}

	seq, err := c.Conn.ModifyBearer(
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
			gtpv2.IFTypeS5S8SGWGTPC,
			options.CplaneSgwIE.Teid,
			options.CplaneSgwIE.IP,
			options.CplaneSgwIE.IP6),
		ie.NewAggregateMaximumBitRate(options.Ambrul, options.Ambrdl),
		ie.NewMobileEquipmentIdentity(options.Mei),
		ie.NewUETimeZone(9, 0),

		ie.NewBearerContext(
			ie.NewEPSBearerID(options.Epsbearerid),
			ie.NewFullyQualifiedTEID(
				gtpv2.IFTypeS5S8SGWGTPU,
				options.UplaneIE.Teid,
				options.UplaneIE.IP,
				options.UplaneIE.IP6,
			).WithInstance(1),
		),
		ie.NewRecovery(0),
	)
	return seq, err
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
		ie.NewPDNType(gtpv2.PDNTypeIPv4),
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

func (c *K6GTPv2Client) CheckSendCreateSessionRequestS5S8(daddr string, options S5S8SgwParams) (bool, error) {
	_, seq, err := c.SendCreateSessionRequestS5S8(daddr, options)
	if err != nil {
		return false, err
	}
	res, err := c.CheckRecvCreateSessionResponse(seq, options.Imsi)
	return res, err
}

func (c *K6GTPv2Client) CheckSendDeleteSessionRequestS5S8(daddr string, options S5S8SgwParams) (bool, error) {
	seq, err := c.SendDeleteSessionRequestS5S8(daddr, options)
	if err != nil {
		return false, err
	}
	return c.CheckRecvDeleteSessionResponse(seq)
}

func (c *K6GTPv2Client) CheckSendModifyBearerRequestS5S8(daddr string, options S5S8SgwParams) (bool, error) {
	seq, err := c.SendModifyBearerRequestS5S8(daddr, options)
	if err != nil {
		return false, err
	}
	return c.CheckRecvModifyBearerResponse(seq)
}

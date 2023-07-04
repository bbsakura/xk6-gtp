package gtpv2

import (
	"fmt"
	"net"
	"strings"

	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/ie"
)

type S5S8SgwParams struct {
	Imsi   string
	Msisdn string
	Mei    string
	// ULI         string
	Mcc           string
	Mnc           string
	Tac           uint16
	Ecgi          string
	Rat           string
	Ftei          string
	Apn           string
	Eci           uint32
	Epsbearerid   uint8
	Uplaneteid    uint32
	Cplanesgwteid uint32
	Cplanepgwteid uint32
	Ambrul        uint32
	Ambrdl        uint32
}

func (c *K6GTPv2Client) SendCreateSessionRequestS5S8(daddr string, options S5S8SgwParams) (*gtpv2.Session, uint32, error) {
	d, err := net.ResolveUDPAddr("udp", daddr)
	if err != nil {
		return nil, 0, fmt.Errorf("resolve udp error")
	}

	localIP := strings.Split(c.Conn.LocalAddr().String(), ":")[0]

	// todo v6
	var cteidIE *ie.IE
	if options.Cplanesgwteid == 0 {
		cteidIE = c.Conn.NewSenderFTEID(localIP, "")
	} else {
		// todo register teid
		cteidIE = ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8SGWGTPC, options.Cplanesgwteid, localIP, "")
		// c.Conn.RegisterSession(options.Cplaneteid, nil)
	}
	uteidIE := ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8SGWGTPU, options.Uplaneteid, localIP, "").WithInstance(2) //dummy uplane teid
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
		d, err := net.ResolveUDPAddr("udp", daddr)
		if err != nil {
			return 0, fmt.Errorf("resolve udp error")
		}
		localIP := strings.Split(c.Conn.LocalAddr().String(), ":")[0]
		cteidIE := ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8SGWGTPC, options.Cplanesgwteid, localIP, "")
		uteidIE := ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8SGWGTPU, options.Uplaneteid, localIP, "").WithInstance(2) //dummy uplane teid

		s5Session, err = c.Conn.ParseCreateSession(d, c.genS5S8SessionIE(options, cteidIE, uteidIE, localIP)...)
		if err != nil {
			return 0, err
		}
		s5Session.AddTEID(uteidIE.MustInterfaceType(), uteidIE.MustTEID())
		c.Conn.RegisterSession(cteidIE.MustTEID(), s5Session)
	} else {
		s5Session, err = c.Conn.GetSessionByIMSI(options.Imsi)
		if err != nil {
			return 0, err
		}
		// teid override
		if options.Cplanepgwteid == 0 {
			options.Cplanepgwteid, err = s5Session.GetTEID(gtpv2.IFTypeS5S8PGWGTPC)
			if err != nil {
				return 0, err
			}
		}
	}
	seq, err := c.Conn.DeleteSession(
		options.Cplanepgwteid,
		s5Session,
		ie.NewEPSBearerID(options.Epsbearerid),
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

package gtpv2

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/ie"
	"github.com/wmnsk/go-gtp/gtpv2/message"
	"go.k6.io/k6/js/modules"
)

const version = "v0.0.1"

type RootModule struct{}

var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &K6GTPv2{}
)

type K6GTPv2 struct {
	Version   string
	vu        modules.VU
	GenParams GenerateBaseParams
	Conn      *gtpv2.Conn
}

// NewModuleInstance implements the modules.Module interface to return
// a new instance for each VU.
func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &K6GTPv2{
		vu:      vu,
		Version: version,
	}
}

// Exports implements the modules.Instance interface and returns the exports
// of the JS module.
func (c *K6GTPv2) Exports() modules.Exports {
	return modules.Exports{Default: c}
}

type ConnectionOptions struct {
	Saddr      string
	Daddr      string
	Count      int
	IFTypeName string
}

func (c *K6GTPv2) Connect(options ConnectionOptions) (*gtpv2.Conn, error) {
	var err error
	saddr, err := net.ResolveUDPAddr("udp", options.Saddr)
	if err != nil {
		return nil, errors.WithMessage(err, "resolve udp error")
	}

	daddr, err := net.ResolveUDPAddr("udp", options.Daddr)
	if err != nil {
		return nil, errors.WithMessage(err, "resolve udp error")
	}

	iftype := IFTypeS11MMEGTPC
	if options.IFTypeName != "" {
		iftype, err = EnumIFTypeString(options.IFTypeName)
		if err != nil {
			return nil, errors.WithMessage(err, "invalid IFTypeName")
		}
	}

	conn, err := gtpv2.Dial(
		context.Background(),
		saddr,
		daddr,
		uint8(iftype),
		uint8(options.Count),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to gtpv2 dial")
	}
	c.Conn = conn

	// ie := map[uint8]gtp2c.InfomationElement{
	// 	gtp2c.IMSIType:           gtp2c.NewIMSIIE(generateIMSI(imsiTail)),
	// 	gtp2c.MSISDNType:         gtp2c.NewMSISDNIE([]uint8{8, 5, 2, 0, 0, 0, 1, 1, 1, 1, 1, 1}),
	// 	gtp2c.MEIType:            gtp2c.NewMEIIE([]uint8{9, 9, 9, 9, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0}),
	// 	gtp2c.ULIType:            gtp2c.NewULIIE(1, 1, &gtp2c.TAI{MCC: 123, MNC: 45, TAC: 12345}, &gtp2c.ECGI{MCC: 123, MNC: 45, ECI: 1234567890}),
	// 	gtp2c.ServingNetworkType: gtp2c.NewServingNetworkIE(440, 20), //
	// 	gtp2c.RATTypeType:        gtp2c.NewRATTypeIE(6),
	// 	gtp2c.FTEIDType:          gtp2c.NewFTEIDIE(1, 1, gtp2c.S5S8PGWGTPC, cteid, net.IP{103, 95, 103, 107}),
	// 	gtp2c.APNType:            gtp2c.NewAPNIE("bbsdev"),
	// 	gtp2c.SelectionModeType:  gtp2c.NewSelectionModeIE(0),
	// 	gtp2c.PDNTypeType:        gtp2c.NewPDNTypeIE(gtp2c.PDNTypeIPv4),
	// 	gtp2c.PAAType:            gtp2c.NewPAAIE(gtp2c.PDNTypeIPv4, net.IP{0, 0, 0, 0}),
	// 	gtp2c.APNRestrictionType: gtp2c.NewAPNRestrictionIE(0), // No Existing Contexts or Restriction
	// 	gtp2c.AMBRType:           gtp2c.NewAMBRIE(1024, 1024),
	// 	gtp2c.BearerContextType: gtp2c.NewBearerContextIE(map[uint8]gtp2c.InfomationElement{
	// 		gtp2c.EBIType:       gtp2c.NewEBIIE(1),
	// 		gtp2c.FTEIDType:     gtp2c.NewFTEIDIE(2, 1, gtp2c.S5S8PGWGTPU, uteid, net.IP{103, 95, 103, 107}),
	// 		gtp2c.BearerQoSType: gtp2c.NewBearerQoSIE(0, 1, 0, 9, 0, 0, 0, 0),
	// 	}),
	// 	gtp2c.RecoveryType: gtp2c.NewRecoveryIE(0),
	// }

	c.GenParams = GenerateBaseParams{
		IMSI:   "454060000000000",
		MSISDN: "852000111111",
		MEI:    "9999000000000101",
		ULI:    "9999000000000101",
	}
	return conn, nil
}

func (c *K6GTPv2) SendEchoRequest(conn *gtpv2.Conn, daddr string) (uint32, error) {
	d, err := net.ResolveUDPAddr("udp", daddr)
	if err != nil {
		return 0, errors.WithMessage(err, "resolve udp error")
	}
	return conn.EchoRequest(d)
}

func (c *K6GTPv2) SendCreateSessionRequest(daddr string, ie ...*ie.IE) (*gtpv2.Session, uint32, error) {
	d, err := net.ResolveIPAddr("ip", daddr)
	if err != nil {
		return nil, 0, fmt.Errorf("resolve ip error")
	}
	sess, seq, err := c.Conn.CreateSession(d, ie...)

	return sess, seq, err
}

type S5S8SgwParams struct {
	IMSI   string
	MSISDN string
	MEI    string
	ULI    string
	MCC    string
	MNC    string
	TAC    string
	ECGI   string
	RAT    string
	FTEI   string
	APN    string
}

type GenerateBaseParams struct {
	IMSI   string
	MSISDN string
	MEI    string
	ULI    string
	MCC    string
	MNC    string
	TAC    string
	ECGI   string
	RAT    string
	FTEI   string
	APN    string
}

func (c *K6GTPv2) SendCreateSessionRequestS5S8(daddr string, options S5S8SgwParams) (*gtpv2.Session, error) {
	d, err := net.ResolveIPAddr("ip", daddr)
	if err != nil {
		return nil, fmt.Errorf("resolve ip error")
	}

	sess, _, err := c.Conn.CreateSession(d,
		ie.NewIMSI("454060000000000"),
	)

	return sess, err
}

// func generateIMSI(n int) []uint8 {
// 	imsi := []uint8{4, 5, 4, 0, 6, 0, 0, 0, 0, 0,
// 		uint8((n / 10000) % 10),
// 		uint8((n / 1000) % 10),
// 		uint8((n / 100) % 10),
// 		uint8((n / 10) % 10),
// 		uint8(n % 10)}
// 	return imsi
// }

// func (c *K6GTPv2) generateS5S8SgwParams(options S5S8SgwParams) S5S8SgwParams {
// 	// if options.IMSI != "" {
// 	// 	options.
// 	// }
// 	// todo
// 	return options
// }

func (c *K6GTPv2) CheckSendEchoRequestWithReturnResponse(conn *gtpv2.Conn, daddr string) (bool, error) {
	if _, err := c.SendEchoRequest(conn, daddr); err != nil {
		return false, err
	}
	return c.CheckRecvEchoResponse(conn)
}

func (c *K6GTPv2) CheckRecvEchoResponse(conn *gtpv2.Conn) (bool, error) {
	var err error
	buf := make([]byte, 1500)

	if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return false, err
	}
	n, _, err := conn.ReadFrom(buf)
	if err != nil {
		return false, err
	}
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return false, err
	}

	msg, err := message.Parse(buf[:n])
	if err != nil {
		return false, err
	}

	if _, ok := msg.(*message.EchoResponse); !ok {
		return false, &gtpv2.UnexpectedTypeError{Msg: msg}
	}
	return true, nil
}

func (c *K6GTPv2) Close() error {
	err := c.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}

package gtpv2

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/pkg/errors"
	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/ie"
	"github.com/wmnsk/go-gtp/gtpv2/message"
	"go.k6.io/k6/js/modules"
)

const version = "v0.0.1"

type (
	// RootModule is the global module instance that will create module
	// instances for each VU.
	RootModule struct{}

	// ModuleInstance represents an instance of the GRPC module for every VU.
	ModuleInstance struct {
		Version string
		vu      modules.VU
		exports map[string]interface{}
	}
)

var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &ModuleInstance{}
)

// NewModuleInstance implements the modules.Module interface to return
// a new instance for each VU.
func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	mi := &ModuleInstance{
		Version: version,
		vu:      vu,
		exports: make(map[string]interface{}),
	}

	mi.exports["K6GTPv2Client"] = mi.NewK6GTPv2Client
	return mi
}

// Exports implements the modules.Instance interface and returns the exports
// of the JS module.
func (mi *ModuleInstance) Exports() modules.Exports {
	return modules.Exports{
		Named: mi.exports,
	}
}

type ConnectionOptions struct {
	Saddr      string
	Daddr      string
	Count      int
	IFTypeName string
}

type K6GTPv2Client struct {
	vu        modules.VU
	GenParams GenerateBaseParams
	Conn      *gtpv2.Conn
	sessions  *sync.Map
}

// NewClient is the JS constructor for the grpc Client.
func (c *ModuleInstance) NewK6GTPv2Client(call goja.ConstructorCall) *goja.Object {
	rt := c.vu.Runtime()
	cli := &K6GTPv2Client{
		vu:       c.vu,
		sessions: &sync.Map{},
	}
	return rt.ToValue(cli).ToObject(rt)
}

func (c *K6GTPv2Client) Connect(options ConnectionOptions) (bool, error) {
	var err error
	saddr, err := net.ResolveUDPAddr("udp", options.Saddr)
	if err != nil {
		return false, errors.WithMessage(err, "resolve udp error")
	}

	daddr, err := net.ResolveUDPAddr("udp", options.Daddr)
	if err != nil {
		return false, errors.WithMessage(err, "resolve udp error")
	}

	iftype := IFTypeS11MMEGTPC
	if options.IFTypeName != "" {
		iftype, err = EnumIFTypeString(options.IFTypeName)
		if err != nil {
			return false, errors.WithMessage(err, "invalid IFTypeName")
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
		return false, errors.WithMessage(err, "failed to gtpv2 dial")
	}
	setHandlers(conn, c.sessions)

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
	c.Conn = conn
	c.GenParams = GenerateBaseParams{
		IMSI:   "454060000000000",
		MSISDN: "852000111111",
		MEI:    "9999000000000101",
		ULI:    "9999000000000101",
	}
	return false, nil
}

func GetMessage[PT *T, T any](ctx context.Context, sessions *sync.Map, msgType uint8, seq uint32) (PT, error) {
	for {
		// TODO reduce cpu usage
		select {
		case <-ctx.Done():
			return (PT)(nil), ctx.Err()
		default:
			if msg, ok := sessions.Load(sessionKey{MessageType: msgType, Sequence: seq}); ok {
				return msg.(PT), nil
			}
		}
	}
}

func setHandlers(conn *gtpv2.Conn, sessions *sync.Map) {
	conn.AddHandler(message.MsgTypeEchoResponse, GetHandler(sessions, message.MsgTypeEchoResponse))
}

type sessionKey struct {
	MessageType uint8
	Sequence    uint32
}

func GetHandler(dst *sync.Map, msgType uint8) func(c *gtpv2.Conn, senderAddr net.Addr, msg message.Message) error {
	return func(c *gtpv2.Conn, senderAddr net.Addr, msg message.Message) error {
		dst.Store(sessionKey{MessageType: msgType, Sequence: msg.Sequence()}, msg)
		return nil
	}
}

func (c *K6GTPv2Client) Exports() *goja.Object {
	rt := c.vu.Runtime()

	return rt.ToValue(c).ToObject(rt)
}

func (c *K6GTPv2Client) SendEchoRequest(daddr string) (uint32, error) {
	d, err := net.ResolveUDPAddr("udp", daddr)
	if err != nil {
		return 0, errors.WithMessage(err, "resolve udp error")
	}
	return c.Conn.EchoRequest(d)
}

func (c *K6GTPv2Client) SendCreateSessionRequest(daddr string, ie ...*ie.IE) (*gtpv2.Session, uint32, error) {
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

func (c *K6GTPv2Client) SendCreateSessionRequestS5S8(daddr string, options S5S8SgwParams) (*gtpv2.Session, error) {
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

func (c *K6GTPv2Client) CheckSendEchoRequestWithReturnResponse(daddr string) (bool, error) {
	seq, err := c.SendEchoRequest(daddr)
	if err != nil {
		return false, err
	}
	return c.CheckRecvEchoResponse(seq)
}

func (c *K6GTPv2Client) CheckRecvEchoResponse(seq uint32) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := GetMessage[*message.EchoResponse](ctx, c.sessions, message.MsgTypeEchoResponse, seq)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *K6GTPv2Client) Close() error {
	err := c.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}

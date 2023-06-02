package gtpv2

import (
	"context"
	"fmt"
	"net"
	"strings"
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
	vu       modules.VU
	Conn     *gtpv2.Conn
	sessions *sync.Map
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

	c.Conn = conn

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
	d, err := net.ResolveIPAddr("udp", daddr)
	if err != nil {
		return nil, 0, fmt.Errorf("resolve udp error")
	}
	sess, seq, err := c.Conn.CreateSession(d, ie...)

	return sess, seq, err
}

type S5S8SgwParams struct {
	IMSI        string
	MSISDN      string
	MEI         string
	ULI         string
	MCC         string
	MNC         string
	TAC         uint16
	ECGI        string
	RAT         string
	FTEI        string
	APN         string
	ECI         uint32
	EPSBearerID uint8
	UplaneTEID  uint32
	AMBRUL      uint32
	AMBRDL      uint32
}

func (c *K6GTPv2Client) SendCreateSessionRequestS5S8(daddr string, options S5S8SgwParams) (*gtpv2.Session, error) {
	d, err := net.ResolveIPAddr("udp", daddr)
	if err != nil {
		return nil, fmt.Errorf("resolve udp error")
	}

	localIP := strings.Split(c.Conn.LocalAddr().String(), ":")[0]

	sess, _, err := c.Conn.CreateSession(d,
		ie.NewIMSI(options.IMSI),
		ie.NewMSISDN(options.MSISDN),
		ie.NewMobileEquipmentIdentity(options.MEI),
		ie.NewAccessPointName(options.APN),
		ie.NewServingNetwork(options.MCC, options.MNC),
		ie.NewRATType(gtpv2.RATTypeEUTRAN),
		c.Conn.NewSenderFTEID(localIP, ""), // todo v6
		ie.NewSelectionMode(gtpv2.SelectionModeMSorNetworkProvidedAPNSubscribedVerified),
		ie.NewPDNType(gtpv2.PDNTypeIPv4),
		ie.NewPDNAddressAllocation("0.0.0.0"),
		ie.NewAPNRestriction(gtpv2.APNRestrictionPublic2),
		ie.NewAggregateMaximumBitRate(options.AMBRUL, options.AMBRDL),
		ie.NewUserLocationInformationStruct(
			nil, nil, nil,
			ie.NewTAI(options.MCC, options.MNC, options.TAC),
			ie.NewECGI(options.MCC, options.MNC, options.ECI),
			nil, nil, nil,
		),
		ie.NewBearerContext(
			ie.NewEPSBearerID(options.EPSBearerID),
			ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8PGWGTPU, options.UplaneTEID, localIP, "").WithInstance(1), //dummy uplane teid
			ie.NewBearerQoS(1, 2, 1, 0xff, 0, 0, 0, 0),
		),
		ie.NewFullyQualifiedCSID(localIP, 1).WithInstance(1),
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

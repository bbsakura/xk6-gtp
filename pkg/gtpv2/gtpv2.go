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
	RootModule struct {
		dialPool *sync.Map
		mu       sync.Mutex
	}

	// ModuleInstance represents an instance of the GRPC module for every VU.
	ModuleInstance struct {
		Version string
		vu      modules.VU
		exports map[string]interface{}
		rm      *RootModule
	}
)

var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &ModuleInstance{}
)

func New() *RootModule {
	fmt.Println("New RootModule")
	return &RootModule{
		dialPool: new(sync.Map),
	}
}

// NewModuleInstance implements the modules.Module interface to return
// a new instance for each VU.
func (r *RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	mi := &ModuleInstance{
		Version: version,
		vu:      vu,
		exports: make(map[string]interface{}),
		rm:      r,
	}

	mi.exports["K6GTPv2Client"] = mi.NewK6GTPv2Client
	mi.exports["K6GTPv2ClientWithConnect"] = mi.NewK6GTPv2ClientWithConnect
	mi.exports["GenerateDummyIMSI"] = GenerateDummyIMSI
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
	Saddr      string `json:"saddr"`
	Daddr      string `json:"daddr"`
	Count      int    `json:"count"`
	IfTypeName string `json:"if_type_name"`
}

type K6GTPv2Client struct {
	vu       modules.VU
	Conn     *gtpv2.Conn
	sessions *sync.Map
	timeout  int64
}

// NewClient is the JS constructor for the grpc Client.
func (c *ModuleInstance) NewK6GTPv2Client(call goja.ConstructorCall) *goja.Object {
	cli := &K6GTPv2Client{
		vu:       c.vu,
		sessions: &sync.Map{},
		timeout:  3,
	}
	rt := c.vu.Runtime()
	return rt.ToValue(cli).ToObject(rt)
}

func (c *ModuleInstance) NewK6GTPv2ClientWithConnect(call goja.ConstructorCall) *goja.Object {
	c.rm.mu.Lock()
	defer c.rm.mu.Unlock()
	op := call.Arguments[0].Export()
	options, err := MapToConnectionOptions(op.(map[string]interface{}))
	if err != nil {
		panic(err)
	}
	cli := c.rm.connGetPool(options.Saddr)
	if cli == nil {
		cli = &K6GTPv2Client{
			vu:       c.vu,
			sessions: &sync.Map{},
			timeout:  3,
		}
		_, err := cli.Connect(options)
		if err != nil {
			panic(err)
		}
		c.rm.connSetPool(options.Saddr, cli)
	}
	rt := c.vu.Runtime()
	return rt.ToValue(cli).ToObject(rt)
}

func (c *RootModule) connSetPool(saddr string, gtpv2 *K6GTPv2Client) {
	c.dialPool.Store(saddr, gtpv2)
}

func (c *RootModule) connGetPool(saddr string) *K6GTPv2Client {
	if gtpv2, ok := c.dialPool.Load(saddr); ok {
		return gtpv2.(*K6GTPv2Client)
	}
	return nil
}

func MapToConnectionOptions(m map[string]interface{}) (ConnectionOptions, error) {
	var co ConnectionOptions
	var ok bool
	if countFloat, ok := m["count"].(int64); ok {
		co.Count = int(countFloat)
	} else {
		return ConnectionOptions{}, fmt.Errorf("count must be a number, but was %T", m["count"])
	}

	if co.Daddr, ok = m["daddr"].(string); !ok {
		return ConnectionOptions{}, fmt.Errorf("daddr must be a string, but was %T", m["daddr"])
	}

	if co.Saddr, ok = m["saddr"].(string); !ok {
		return ConnectionOptions{}, fmt.Errorf("saddr must be a string, but was %T", m["saddr"])
	}

	if co.IfTypeName, ok = m["if_type_name"].(string); !ok {
		return ConnectionOptions{}, fmt.Errorf("if_type_name must be a string, but was %T", m["if_type_name"])
	}

	return co, nil
}

func (c *K6GTPv2Client) Connect(options ConnectionOptions) (bool, error) {
	if c.Conn != nil {
		return true, nil // already connected
	}
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
	if options.IfTypeName != "" {
		iftype, err = EnumIFTypeString(options.IfTypeName)
		if err != nil {
			return false, errors.WithMessage(err, "invalid IfTypeName")
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

func (c *K6GTPv2Client) SetTimeout(timeout int64) {
	c.timeout = timeout
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
	conn.AddHandler(message.MsgTypeCreateSessionResponse, GetHandler(sessions, message.MsgTypeCreateSessionResponse))
	conn.AddHandler(message.MsgTypeDeleteSessionResponse, GetHandler(sessions, message.MsgTypeDeleteSessionResponse))
	conn.AddHandler(message.MsgTypeModifyBearerResponse, GetHandler(sessions, message.MsgTypeModifyBearerResponse))
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

func (c *K6GTPv2Client) CheckSendEchoRequestWithReturnResponse(daddr string) (bool, error) {
	seq, err := c.SendEchoRequest(daddr)
	if err != nil {
		return false, err
	}
	return c.CheckRecvEchoResponse(seq)
}

func (c *K6GTPv2Client) CheckRecvEchoResponse(seq uint32) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeout)*time.Second)
	defer cancel()
	_, err := GetMessage[*message.EchoResponse](ctx, c.sessions, message.MsgTypeEchoResponse, seq)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *K6GTPv2Client) CheckRecvCreateSessionResponse(seq uint32, imsi string) (bool, error) {
	sess, err := c.Conn.GetSessionByIMSI(imsi)
	if err != nil {
		return false, err
	}
	fmt.Println(c.timeout)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeout)*time.Second)
	defer cancel()
	res, err := GetMessage[*message.CreateSessionResponse](ctx, c.sessions, message.MsgTypeCreateSessionResponse, seq)
	if err != nil {
		return false, err
	}
	if fteidcIE := res.PGWS5S8FTEIDC; fteidcIE != nil {
		it, err := fteidcIE.InterfaceType()
		if err != nil {
			return true, nil
		}
		teid, err := fteidcIE.TEID()
		if err != nil {
			return true, nil
		}
		sess.AddTEID(it, teid)
	}
	return true, nil
}

func (c *K6GTPv2Client) CheckRecvDeleteSessionResponse(seq uint32) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeout)*time.Second)
	defer cancel()
	_, err := GetMessage[*message.DeleteSessionResponse](ctx, c.sessions, message.MsgTypeDeleteSessionResponse, seq)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *K6GTPv2Client) CheckRecvModifyBearerResponse(seq uint32) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeout)*time.Second)
	defer cancel()
	_, err := GetMessage[*message.ModifyBearerResponse](ctx, c.sessions, message.MsgTypeModifyBearerResponse, seq)
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

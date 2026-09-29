package gtpv2

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/grafana/sobek"
	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/ie"
	"github.com/wmnsk/go-gtp/gtpv2/message"
	"go.k6.io/k6/js/modules"
)

const version = "v0.0.1"

// errRunOnly is returned when a JS-callable method that requires VU run state
// (samples, tags) is invoked from the init context.
var errRunOnly = errors.New("gtpv2: this operation is only available in the run stage (VU state is nil)")

type (
	// RootModule is the global module instance that will create module
	// instances for each VU.
	RootModule struct {
		dialPool    *sync.Map
		mu          sync.Mutex
		metricsOnce sync.Once
		metrics     *gtpMetrics
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
	return &RootModule{
		dialPool: new(sync.Map),
	}
}

// NewModuleInstance implements the modules.Module interface to return
// a new instance for each VU.
func (r *RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	r.metricsOnce.Do(func() {
		if env := vu.InitEnv(); env != nil && env.Registry != nil {
			r.metrics = registerMetrics(env.Registry)
		}
	})

	mi := &ModuleInstance{
		Version: version,
		vu:      vu,
		exports: make(map[string]interface{}),
		rm:      r,
	}

	mi.exports["K6GTPv2Client"] = mi.NewK6GTPv2Client
	mi.exports["K6GTPv2ClientWithConnect"] = mi.NewK6GTPv2ClientWithConnect
	mi.exports["K6GTPv2Responder"] = mi.NewK6GTPv2Responder
	mi.exports["GenerateDummyIMSI"] = GenerateDummyIMSI
	mi.exports["ie"] = ieExports()
	mi.exports["msg"] = msgExports()
	mi.exports["IFType"] = ifTypeExports()
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
	rm       *RootModule
	vu       modules.VU
	Conn     *gtpv2.Conn
	sessions *sync.Map
	timeout  int64

	// connCtx bounds the underlying gtpv2.Conn's read loop lifetime. It is
	// created lazily in Connect and cancelled by Close. Deliberately not
	// tied to modules.VU.Context() — that context is cancelled at each
	// iteration boundary, which would close the socket after the first
	// iteration and break subsequent sends.
	connCtx    context.Context
	connCancel context.CancelFunc
}

// NewClient is the JS constructor for the grpc Client.
func (c *ModuleInstance) NewK6GTPv2Client(_ sobek.ConstructorCall) *sobek.Object {
	cli := &K6GTPv2Client{
		rm:       c.rm,
		vu:       c.vu,
		sessions: &sync.Map{},
		timeout:  3,
	}
	rt := c.vu.Runtime()
	return rt.ToValue(cli).ToObject(rt)
}

func (c *ModuleInstance) NewK6GTPv2ClientWithConnect(call sobek.ConstructorCall) *sobek.Object {
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
			rm:       c.rm,
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

// requireRunState returns errRunOnly when invoked outside the run stage so
// init-context callers get a clear error instead of a nil deref inside k6.
func (c *K6GTPv2Client) requireRunState() error {
	if c.vu == nil || c.vu.State() == nil {
		return errRunOnly
	}
	return nil
}

// iterCtx returns the per-iteration context if the VU is running, otherwise
// falls back to Background. Callers must not cache the returned value.
func (c *K6GTPv2Client) iterCtx() context.Context {
	if c.vu == nil {
		return context.Background()
	}
	if ctx := c.vu.Context(); ctx != nil {
		return ctx
	}
	return context.Background()
}

func (c *K6GTPv2Client) timeoutDuration() time.Duration {
	if c.timeout <= 0 {
		return 3 * time.Second
	}
	return time.Duration(c.timeout) * time.Second
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
	if err := c.requireRunState(); err != nil {
		return false, err
	}
	saddr, err := net.ResolveUDPAddr("udp", options.Saddr)
	if err != nil {
		return false, fmt.Errorf("resolve source UDP addr %q: %w", options.Saddr, err)
	}

	daddr, err := net.ResolveUDPAddr("udp", options.Daddr)
	if err != nil {
		return false, fmt.Errorf("resolve destination UDP addr %q: %w", options.Daddr, err)
	}

	iftype := IFTypeS11MMEGTPC
	if options.IfTypeName != "" {
		iftype, err = EnumIFTypeString(options.IfTypeName)
		if err != nil {
			return false, fmt.Errorf("invalid IfTypeName %q: %w", options.IfTypeName, err)
		}
	}

	if options.Count < 0 || options.Count > 255 {
		return false, fmt.Errorf("count %d out of range [0, 255] for GTPv2 Restart Counter", options.Count)
	}

	// gtpv2.Dial's receive loop lives as long as the ctx we pass in, so bind
	// it to a Client-scoped context rather than the per-iteration one. Close
	// cancels it explicitly.
	c.connCtx, c.connCancel = context.WithCancel(context.Background())
	// #nosec G115 -- options.Count is bounded to [0, 255] by the check above.
	conn, err := gtpv2.Dial(
		c.connCtx,
		saddr,
		daddr,
		uint8(iftype),
		uint8(options.Count),
	)
	if err != nil {
		c.connCancel()
		c.connCancel = nil
		return false, fmt.Errorf("gtpv2 dial %s -> %s: %w", saddr, daddr, err)
	}
	// go-gtp starts the receive loop inside Dial; register handlers immediately
	// so an early reply cannot slip through unhandled.
	setHandlers(conn, c.sessions)

	c.Conn = conn
	return true, nil
}

func (c *K6GTPv2Client) SetTimeout(timeout int64) {
	c.timeout = timeout
}

// pendingEntry rendezvous slot between a request sender waiting for a reply and
// the receive-side handler that delivers it. One of msg/ch is set depending on
// which side reached the sessions map first; see waitForMessage and
// storeMessageHandler.
type pendingEntry struct {
	msg message.Message
	ch  chan struct{}
}

// GetMessage waits for a GTPv2 response matching (msgType, seq) via a channel
// rendezvous. Kept as a generic wrapper so the typed CheckRecv* helpers stay
// concise; no CPU is burned while waiting.
func GetMessage[PT *T, T any](ctx context.Context, sessions *sync.Map, msgType uint8, seq uint32) (PT, error) {
	msg, err := waitForMessage(ctx, sessions, msgType, seq)
	if err != nil {
		return PT(nil), err
	}
	typed, ok := msg.(PT)
	if !ok {
		return PT(nil), fmt.Errorf("gtpv2: unexpected message type %T for seq=%d", msg, seq)
	}
	return typed, nil
}

func waitForMessage(ctx context.Context, sessions *sync.Map, msgType uint8, seq uint32) (message.Message, error) {
	key := sessionKey{MessageType: msgType, Sequence: seq}
	waiter := &pendingEntry{ch: make(chan struct{})}
	actual, loaded := sessions.LoadOrStore(key, waiter)
	if loaded {
		// Handler beat us — take the buffered message and clear the slot.
		sessions.Delete(key)
		return actual.(*pendingEntry).msg, nil
	}
	select {
	case <-waiter.ch:
		sessions.Delete(key)
		return waiter.msg, nil
	case <-ctx.Done():
		// Best-effort cleanup. If the handler races in between here and
		// Delete, it will overwrite with its own buffered entry which then
		// stays until the next call for the same key.
		sessions.Delete(key)
		return nil, ctx.Err()
	}
}

func setHandlers(conn *gtpv2.Conn, sessions *sync.Map) {
	for _, mt := range []uint8{
		message.MsgTypeEchoResponse,
		message.MsgTypeCreateSessionResponse,
		message.MsgTypeDeleteSessionResponse,
		message.MsgTypeModifyBearerResponse,
	} {
		conn.AddHandler(mt, storeMessageHandler(sessions, mt))
	}
}

type sessionKey struct {
	MessageType uint8
	Sequence    uint32
}

func storeMessageHandler(dst *sync.Map, msgType uint8) func(c *gtpv2.Conn, senderAddr net.Addr, msg message.Message) error {
	return func(_ *gtpv2.Conn, _ net.Addr, msg message.Message) error {
		key := sessionKey{MessageType: msgType, Sequence: msg.Sequence()}
		incoming := &pendingEntry{msg: msg}
		actual, loaded := dst.LoadOrStore(key, incoming)
		if !loaded {
			// No waiter yet — leave the message buffered for the next Load.
			return nil
		}
		waiter := actual.(*pendingEntry)
		waiter.msg = msg
		close(waiter.ch)
		return nil
	}
}

func (c *K6GTPv2Client) SendEchoRequest(daddr string) (uint32, error) {
	d, err := net.ResolveUDPAddr("udp", daddr)
	if err != nil {
		return 0, fmt.Errorf("resolve destination UDP addr %q: %w", daddr, err)
	}
	seq, err := c.Conn.EchoRequest(d)
	if err == nil {
		c.rm.metrics.pushRequest(c.iterCtx(), c.vu.State(), msgTypeEcho)
	}
	return seq, err
}

func (c *K6GTPv2Client) SendCreateSessionRequest(daddr string, ie ...*ie.IE) (*gtpv2.Session, uint32, error) {
	d, err := net.ResolveUDPAddr("udp", daddr)
	if err != nil {
		return nil, 0, fmt.Errorf("resolve destination UDP addr %q: %w", daddr, err)
	}
	sess, seq, err := c.Conn.CreateSession(d, ie...)
	if err == nil {
		c.rm.metrics.pushRequest(c.iterCtx(), c.vu.State(), msgTypeCreateSession)
	}
	return sess, seq, err
}

func (c *K6GTPv2Client) CheckSendEchoRequestWithReturnResponse(daddr string) (bool, error) {
	start := time.Now()
	seq, err := c.SendEchoRequest(daddr)
	if err != nil {
		return false, err
	}
	ok, err := c.CheckRecvEchoResponse(seq)
	c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeEcho, time.Since(start), err != nil)
	return ok, err
}

// TryEcho sends an Echo Request and waits for the paired response, returning a
// SendResult with the sequence, whether it succeeded, the elapsed round-trip
// time, and — on timeout — a Timeout=true flag. Echo does not carry a Cause IE
// so Cause is always 0 on success.
func (c *K6GTPv2Client) TryEcho(daddr string) *SendResult {
	start := time.Now()
	seq, err := c.SendEchoRequest(daddr)
	if err != nil {
		return newSendError(time.Since(start), err)
	}
	_, recvErr := c.CheckRecvEchoResponse(seq)
	elapsed := time.Since(start)
	c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeEcho, elapsed, recvErr != nil)
	if recvErr != nil {
		r := newSendError(elapsed, recvErr)
		r.Sequence = seq
		return r
	}
	return newSendOK(seq, 0, elapsed)
}

// SendRaw dispatches an arbitrary GTPv2 message (typically built via
// gtpv2.msg.*) to daddr and returns a SendHandle with the assigned sequence.
// Pair the sequence with AwaitMessage to inspect the response. Cause and
// timing are only recorded when the JS side later awaits — request_total is
// emitted here so the counter still fires when the response is intentionally
// dropped (abnormal-path tests).
func (c *K6GTPv2Client) SendRaw(daddr string, msg message.Message) *SendHandle {
	d, err := net.ResolveUDPAddr("udp", daddr)
	if err != nil {
		return &SendHandle{Error: fmt.Errorf("resolve destination UDP addr %q: %w", daddr, err).Error()}
	}
	seq, err := c.Conn.SendMessageTo(msg, d)
	if err != nil {
		return &SendHandle{Error: err.Error()}
	}
	c.rm.metrics.pushRequest(c.iterCtx(), c.vu.State(), messageTypeTag(msg.MessageType()))
	return &SendHandle{Ok: true, Sequence: seq}
}

// AwaitMessage waits up to timeoutMs milliseconds for a GTPv2 message with the
// given type and sequence number. Extracts the Cause IE when the message type
// carries one; otherwise Cause is 0. A handler for msgType is registered
// idempotently on the underlying gtpv2.Conn so callers do not have to know in
// advance which message types to hook up.
func (c *K6GTPv2Client) AwaitMessage(msgType uint8, seq uint32, timeoutMs int64) *AwaitResult {
	c.Conn.AddHandler(msgType, storeMessageHandler(c.sessions, msgType))
	timeout := time.Duration(timeoutMs) * time.Millisecond
	if timeoutMs <= 0 {
		timeout = c.timeoutDuration()
	}
	ctx, cancel := context.WithTimeout(c.iterCtx(), timeout)
	defer cancel()
	start := time.Now()
	msg, err := waitForMessage(ctx, c.sessions, msgType, seq)
	elapsed := time.Since(start)
	if err != nil {
		return newAwaitError(elapsed, err)
	}
	return newAwaitOK(extractCause(msg), elapsed)
}

// messageTypeTag maps a GTPv2 numeric message type to the low-cardinality tag
// value used for the metric emissions. Unknown types get their decimal code
// so the metric never carries an unbounded string.
func messageTypeTag(t uint8) string {
	switch t {
	case message.MsgTypeEchoRequest:
		return msgTypeEcho
	case message.MsgTypeCreateSessionRequest:
		return msgTypeCreateSession
	case message.MsgTypeDeleteSessionRequest:
		return msgTypeDeleteSession
	case message.MsgTypeModifyBearerRequest:
		return msgTypeModifyBearer
	default:
		return causeString(t, true) // reuse the numeric formatter
	}
}

// extractCause pulls the numeric Cause IE value from any message that carries
// one, returning 0 when the message does not have Cause.
func extractCause(msg message.Message) uint8 {
	if ie := responseCauseIE(msg); ie != nil {
		if v, err := ie.Cause(); err == nil {
			return v
		}
	}
	return 0
}

// responseCauseIE returns the Cause slot IE for the response message types
// this extension knows about. Returns nil when the message does not carry
// Cause. Kept as a type switch so we do not depend on go-gtp exposing a Cause
// accessor method.
func responseCauseIE(msg message.Message) *ie.IE {
	switch m := msg.(type) {
	case *message.CreateSessionResponse:
		return m.Cause
	case *message.DeleteSessionResponse:
		return m.Cause
	case *message.ModifyBearerResponse:
		return m.Cause
	case *message.CreateBearerResponse:
		return m.Cause
	case *message.UpdateBearerResponse:
		return m.Cause
	case *message.DeleteBearerResponse:
		return m.Cause
	}
	return nil
}

// TryCreateSessionRaw sends a Create Session Request built from a
// user-supplied list of IEs (typically composed via gtpv2.ie.*) and waits for
// the paired response. imsi is required to look up the session on the receive
// side; it must be present in the IE list.
//
// Use this instead of TryCreateSessionS5S8 when you need to omit, mutate, or
// add IEs beyond what S5S8SgwParams exposes — for example to reproduce a
// missing-IE cause code or an unexpected FTEID interface type.
func (c *K6GTPv2Client) TryCreateSessionRaw(daddr, imsi string, ies []*ie.IE) *SendResult {
	start := time.Now()
	_, seq, err := c.SendCreateSessionRequest(daddr, ies...)
	if err != nil {
		c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeCreateSession, time.Since(start), true)
		return newSendError(time.Since(start), err)
	}
	cause, recvErr := c.CheckRecvCreateSessionResponse(seq, imsi)
	elapsed := time.Since(start)
	c.rm.metrics.pushDuration(c.iterCtx(), c.vu.State(), msgTypeCreateSession, elapsed, recvErr != nil)
	if recvErr != nil {
		r := newSendError(elapsed, recvErr)
		r.Sequence = seq
		return r
	}
	return newSendOK(seq, uint8(cause), elapsed)
}

func (c *K6GTPv2Client) recvCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.iterCtx(), c.timeoutDuration())
}

// recordRecvOutcome emits resp_cause on success and timeout_total when the
// wait was cancelled by the deadline.
func (c *K6GTPv2Client) recordRecvOutcome(msgType string, cause EnumIFCause, err error) {
	state := c.vu.State()
	if state == nil {
		return
	}
	ctx := c.iterCtx()
	switch {
	case err == nil:
		if cause == 0 && msgType == msgTypeEcho {
			c.rm.metrics.pushResponse(ctx, state, msgType, causeNone)
			return
		}
		if cause == 0 {
			c.rm.metrics.pushResponse(ctx, state, msgType, causeUnknown)
			return
		}
		c.rm.metrics.pushResponse(ctx, state, msgType, causeString(uint8(cause), true))
	case errors.Is(err, context.DeadlineExceeded):
		c.rm.metrics.pushTimeout(ctx, state, msgType)
	}
}

func (c *K6GTPv2Client) CheckRecvEchoResponse(seq uint32) (bool, error) {
	ctx, cancel := c.recvCtx()
	defer cancel()
	_, err := GetMessage[*message.EchoResponse](ctx, c.sessions, message.MsgTypeEchoResponse, seq)
	c.recordRecvOutcome(msgTypeEcho, 0, err)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *K6GTPv2Client) CheckRecvCreateSessionResponse(seq uint32, imsi string) (EnumIFCause, error) {
	causeRes := EnumIFCause(0)
	sess, err := c.Conn.GetSessionByIMSI(imsi)
	if err != nil {
		return causeRes, err
	}

	ctx, cancel := c.recvCtx()
	defer cancel()
	res, err := GetMessage[*message.CreateSessionResponse](ctx, c.sessions, message.MsgTypeCreateSessionResponse, seq)
	if err != nil {
		c.recordRecvOutcome(msgTypeCreateSession, 0, err)
		return causeRes, err
	}
	if causeIE := res.Cause; causeIE != nil {
		cause, err := causeIE.Cause()
		if err != nil {
			c.recordRecvOutcome(msgTypeCreateSession, 0, nil)
			return causeRes, err
		}
		causeRes = EnumIFCause(cause)
	}
	c.recordRecvOutcome(msgTypeCreateSession, causeRes, nil)
	if fteidcIE := res.PGWS5S8FTEIDC; fteidcIE != nil {
		it, err := fteidcIE.InterfaceType()
		if err != nil {
			return causeRes, nil
		}
		teid, err := fteidcIE.TEID()
		if err != nil {
			return causeRes, nil
		}
		sess.AddTEID(it, teid)
	}
	return causeRes, nil
}

func (c *K6GTPv2Client) CheckRecvDeleteSessionResponse(seq uint32) (EnumIFCause, error) {
	causeRes := EnumIFCause(0)
	ctx, cancel := c.recvCtx()
	defer cancel()
	res, err := GetMessage[*message.DeleteSessionResponse](ctx, c.sessions, message.MsgTypeDeleteSessionResponse, seq)
	if err != nil {
		c.recordRecvOutcome(msgTypeDeleteSession, 0, err)
		return causeRes, err
	}
	if causeIE := res.Cause; causeIE != nil {
		cause, err := causeIE.Cause()
		if err != nil {
			c.recordRecvOutcome(msgTypeDeleteSession, 0, nil)
			return causeRes, err
		}
		causeRes = EnumIFCause(cause)
	}
	c.recordRecvOutcome(msgTypeDeleteSession, causeRes, nil)
	return causeRes, nil
}

func (c *K6GTPv2Client) CheckRecvModifyBearerResponse(seq uint32) (EnumIFCause, error) {
	causeRes := EnumIFCause(0)
	ctx, cancel := c.recvCtx()
	defer cancel()
	res, err := GetMessage[*message.ModifyBearerResponse](ctx, c.sessions, message.MsgTypeModifyBearerResponse, seq)
	if err != nil {
		c.recordRecvOutcome(msgTypeModifyBearer, 0, err)
		return causeRes, err
	}
	if causeIE := res.Cause; causeIE != nil {
		cause, err := causeIE.Cause()
		if err != nil {
			c.recordRecvOutcome(msgTypeModifyBearer, 0, nil)
			return causeRes, err
		}
		causeRes = EnumIFCause(cause)
	}
	c.recordRecvOutcome(msgTypeModifyBearer, causeRes, nil)
	return causeRes, nil
}

func (c *K6GTPv2Client) Close() error {
	if c.connCancel != nil {
		c.connCancel()
		c.connCancel = nil
	}
	if c.Conn == nil {
		return nil
	}
	err := c.Conn.Close()
	c.Conn = nil
	return err
}

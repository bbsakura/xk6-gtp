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
	"go.k6.io/k6/js/promises"
)

const version = "v0.0.1"

// errRunOnly is returned when a JS-callable method that requires VU run state
// (samples, tags) is invoked from the init context.
var errRunOnly = errors.New("gtpv2: this operation is only available in the run stage (VU state is nil)")

var errNotConnected = errors.New("gtpv2: client is not connected (call connect first)")

// dialEchoTimeout matches the liveness check gtpv2.Dial performs.
const dialEchoTimeout = 3 * time.Second

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
	rm      *RootModule
	vu      modules.VU
	Conn    *gtpv2.Conn
	timeout int64

	// box correlates responses for the current Conn only; see responseBox.
	box *responseBox

	// connCtx bounds the underlying gtpv2.Conn's read loop lifetime. It is
	// created lazily in Connect and cancelled by Close. Deliberately not
	// tied to modules.VU.Context() — that context is cancelled at each
	// iteration boundary, which would close the socket after the first
	// iteration and break subsequent sends.
	connCtx    context.Context
	connCancel context.CancelFunc
	// serveDone is closed once Serve has returned, i.e. the socket is
	// released and the local address can be bound again.
	serveDone chan struct{}

	// hasConnected tracks whether Connect has ever completed on this Client,
	// so a subsequent Connect (after Close) can emit reconnect metric rather
	// than the first-connect no-op.
	hasConnected bool
}

// NewClient is the JS constructor for the grpc Client.
func (c *ModuleInstance) NewK6GTPv2Client(_ sobek.ConstructorCall) *sobek.Object {
	cli := &K6GTPv2Client{
		rm:      c.rm,
		vu:      c.vu,
		timeout: 3,
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
			rm:      c.rm,
			vu:      c.vu,
			timeout: 3,
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

// connReady guards every entry point that touches c.Conn.
func (c *K6GTPv2Client) connReady() error {
	if err := c.requireRunState(); err != nil {
		return err
	}
	if c.Conn == nil || c.box == nil {
		return errNotConnected
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

	// gtpv2.Dial is not used because it hides the Serve goroutine, so Close
	// could not wait for the socket to be released before a reconnect binds
	// the same address. The connection lives on a Client-scoped context:
	// vu.Context() ends at every iteration boundary.
	c.connCtx, c.connCancel = context.WithCancel(context.Background())
	// #nosec G115 -- options.Count is bounded to [0, 255] by the check above.
	conn := gtpv2.NewConn(saddr, uint8(iftype), uint8(options.Count))
	if err := conn.Listen(c.connCtx); err != nil {
		c.connCancel()
		c.connCancel = nil
		return false, fmt.Errorf("gtpv2 listen %s: %w", saddr, err)
	}
	box := newResponseBox()
	setHandlers(conn, box)
	serveDone := make(chan struct{})
	go func() {
		defer close(serveDone)
		_ = conn.Serve(c.connCtx)
	}()
	c.Conn, c.box, c.serveDone = conn, box, serveDone

	if err := c.echoHandshake(daddr); err != nil {
		_ = c.shutdown()
		return false, fmt.Errorf("gtpv2 dial %s -> %s: %w", saddr, daddr, err)
	}

	if c.hasConnected {
		c.rm.metrics.pushReconnect(c.iterCtx(), c.vu.State(), options.Daddr)
	}
	c.hasConnected = true
	return true, nil
}

func (c *K6GTPv2Client) echoHandshake(daddr net.Addr) error {
	seq, err := c.Conn.EchoRequest(daddr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.iterCtx(), dialEchoTimeout)
	defer cancel()
	_, err = c.box.wait(ctx, message.MsgTypeEchoResponse, seq)
	return err
}

func (c *K6GTPv2Client) SetTimeout(timeout int64) {
	c.timeout = timeout
}

func (c *K6GTPv2Client) SendEchoRequest(daddr string) (uint32, error) {
	if err := c.connReady(); err != nil {
		c.rm.metrics.pushSendError(c.iterCtx(), c.vu.State(), msgTypeEcho)
		return 0, err
	}
	d, err := net.ResolveUDPAddr("udp", daddr)
	if err != nil {
		c.rm.metrics.pushSendError(c.iterCtx(), c.vu.State(), msgTypeEcho)
		return 0, fmt.Errorf("resolve destination UDP addr %q: %w", daddr, err)
	}
	seq, err := c.Conn.EchoRequest(d)
	if err != nil {
		c.rm.metrics.pushSendError(c.iterCtx(), c.vu.State(), msgTypeEcho)
		return seq, err
	}
	c.rm.metrics.pushRequest(c.iterCtx(), c.vu.State(), msgTypeEcho)
	return seq, nil
}

func (c *K6GTPv2Client) SendCreateSessionRequest(daddr string, ies ...*ie.IE) (*gtpv2.Session, uint32, error) {
	if err := c.connReady(); err != nil {
		c.rm.metrics.pushSendError(c.iterCtx(), c.vu.State(), msgTypeCreateSession)
		return nil, 0, err
	}
	d, err := net.ResolveUDPAddr("udp", daddr)
	if err != nil {
		c.rm.metrics.pushSendError(c.iterCtx(), c.vu.State(), msgTypeCreateSession)
		return nil, 0, fmt.Errorf("resolve destination UDP addr %q: %w", daddr, err)
	}
	// Same as gtpv2.Conn.CreateSession, but the session is registered before
	// dispatch so the reply's TEID passes go-gtp's validation.
	sess, err := c.Conn.ParseCreateSession(d, ies...)
	if err != nil {
		c.rm.metrics.pushSendError(c.iterCtx(), c.vu.State(), msgTypeCreateSession)
		return nil, 0, err
	}
	msg := message.NewCreateSessionRequest(0, 0, ies...)
	undo := c.registerSenderSession(sess, msg.SenderFTEIDC)
	seq, err := c.Conn.SendMessageTo(msg, d)
	if err != nil {
		undo()
		c.rm.metrics.pushSendError(c.iterCtx(), c.vu.State(), msgTypeCreateSession)
		return nil, seq, err
	}
	c.rm.metrics.pushRequest(c.iterCtx(), c.vu.State(), msgTypeCreateSession)
	return sess, seq, nil
}

// registerSenderSession registers sess under the TEID of the Sender F-TEID we
// advertise, which is the TEID the peer puts in its reply header. Returns a
// func that undoes the registration.
func (c *K6GTPv2Client) registerSenderSession(sess *gtpv2.Session, senderFTEID *ie.IE) func() {
	if sess == nil || senderFTEID == nil {
		return func() {}
	}
	teid, err := senderFTEID.TEID()
	if err != nil || teid == 0 {
		return func() {}
	}
	c.Conn.RegisterSession(teid, sess)
	conn := c.Conn
	return func() { conn.RemoveSession(sess) }
}

func (c *K6GTPv2Client) registerRawSession(raddr net.Addr, msg message.Message) func() {
	csr, ok := msg.(*message.CreateSessionRequest)
	if !ok {
		return func() {}
	}
	sess := gtpv2.NewSession(raddr, &gtpv2.Subscriber{Location: &gtpv2.Location{}})
	if csr.IMSI != nil {
		if imsi, err := csr.IMSI.IMSI(); err == nil {
			sess.IMSI = imsi
		}
	}
	return c.registerSenderSession(sess, csr.SenderFTEIDC)
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
// Pair the sequence with AwaitMessage to inspect the response. req_total is
// emitted here so the counter still fires when the response is intentionally
// never awaited (abnormal-path tests); duration and cause are emitted by the
// await, measured from this dispatch.
func (c *K6GTPv2Client) SendRaw(daddr string, msg message.Message) *SendHandle {
	if msg == nil {
		return &SendHandle{Error: "gtpv2: SendRaw requires a message"}
	}
	msgTypeTag := messageTypeTag(msg.MessageType())
	if err := c.connReady(); err != nil {
		c.rm.metrics.pushSendError(c.iterCtx(), c.vu.State(), msgTypeTag)
		return &SendHandle{Error: err.Error()}
	}
	d, err := net.ResolveUDPAddr("udp", daddr)
	if err != nil {
		c.rm.metrics.pushSendError(c.iterCtx(), c.vu.State(), msgTypeTag)
		return &SendHandle{Error: fmt.Errorf("resolve destination UDP addr %q: %w", daddr, err).Error()}
	}
	undo := c.registerRawSession(d, msg)
	sentAt := time.Now()
	seq, err := c.Conn.SendMessageTo(msg, d)
	if err != nil {
		undo()
		c.rm.metrics.pushSendError(c.iterCtx(), c.vu.State(), msgTypeTag)
		return &SendHandle{Error: err.Error()}
	}
	c.box.markSent(seq, sentAt)
	c.rm.metrics.pushRequest(c.iterCtx(), c.vu.State(), msgTypeTag)
	return &SendHandle{Ok: true, Sequence: seq}
}

// AwaitMessage waits up to timeoutMs milliseconds for a GTPv2 message with the
// given type and sequence number. Cause is the Cause IE value, or 0 when the
// message has none. Response types are captured from Connect onwards; any
// other type gets its handler here, so only replies to such types can be
// missed if they arrive before the await.
func (c *K6GTPv2Client) AwaitMessage(msgType uint8, seq uint32, timeoutMs int64) *AwaitResult {
	if err := c.connReady(); err != nil {
		return newAwaitError(0, err)
	}
	c.ensureHandler(msgType)
	timeout := effectiveTimeout(timeoutMs, c.timeoutDuration())
	ctx, cancel := context.WithTimeout(c.iterCtx(), timeout)
	defer cancel()
	return c.awaitAndRecord(ctx, c.box, msgType, seq)
}

// AwaitMessageAsync is the Promise-returning sibling of AwaitMessage. Each
// call spawns a wait goroutine that resolves the Promise once the response
// arrives (or times out), so scripts can Promise.all over multiple in-flight
// requests dispatched from the same Client. Only one wait per (msgType, seq)
// may be pending at a time.
func (c *K6GTPv2Client) AwaitMessageAsync(msgType uint8, seq uint32, timeoutMs int64) *sobek.Promise {
	promise, resolve, reject := promises.New(c.vu)
	if err := c.connReady(); err != nil {
		reject(err)
		return promise
	}
	c.ensureHandler(msgType)
	timeout := effectiveTimeout(timeoutMs, c.timeoutDuration())
	// Snapshot iteration context and box up front: the goroutine must not
	// observe a later Close/Connect swapping them.
	parent := c.iterCtx()
	box := c.box
	go func() {
		ctx, cancel := context.WithTimeout(parent, timeout)
		defer cancel()
		resolve(c.awaitAndRecord(ctx, box, msgType, seq))
	}()
	return promise
}

func (c *K6GTPv2Client) ensureHandler(msgType uint8) {
	if !isPreregistered(msgType) {
		c.Conn.AddHandler(msgType, storeMessageHandler(c.box, msgType))
	}
}

// awaitAndRecord performs the wait and metric emission both sync and async
// paths share. Elapsed time is measured from SendRaw's dispatch when known.
func (c *K6GTPv2Client) awaitAndRecord(ctx context.Context, box *responseBox, msgType uint8, seq uint32) *AwaitResult {
	start := time.Now()
	msg, err := box.wait(ctx, msgType, seq)
	if errors.Is(err, errDuplicateAwait) {
		return newAwaitError(time.Since(start), err)
	}
	if sentAt, ok := box.takeSentAt(seq); ok {
		start = sentAt
	}
	elapsed := time.Since(start)
	tag := messageTypeTag(msgType)
	state := c.vu.State()
	c.rm.metrics.pushDuration(c.iterCtx(), state, tag, elapsed, err != nil)
	switch {
	case err == nil:
		cause, causeTag := responseCause(msg)
		c.rm.metrics.pushResponse(c.iterCtx(), state, tag, causeTag)
		return newAwaitOK(cause, elapsed)
	case errors.Is(err, context.DeadlineExceeded):
		c.rm.metrics.pushTimeout(c.iterCtx(), state, tag)
		return newAwaitError(elapsed, err)
	default:
		return newAwaitError(elapsed, err)
	}
}

func effectiveTimeout(requestedMs int64, fallback time.Duration) time.Duration {
	if requestedMs <= 0 {
		return fallback
	}
	return time.Duration(requestedMs) * time.Millisecond
}

// messageTypeTag maps a GTPv2 numeric message type to the low-cardinality tag
// value used for the metric emissions. A request and its response share one
// tag so thresholds can be written per procedure. Unknown types get their
// decimal code so the metric never carries an unbounded string.
func messageTypeTag(t uint8) string {
	switch t {
	case message.MsgTypeEchoRequest, message.MsgTypeEchoResponse:
		return msgTypeEcho
	case message.MsgTypeCreateSessionRequest, message.MsgTypeCreateSessionResponse:
		return msgTypeCreateSession
	case message.MsgTypeDeleteSessionRequest, message.MsgTypeDeleteSessionResponse:
		return msgTypeDeleteSession
	case message.MsgTypeModifyBearerRequest, message.MsgTypeModifyBearerResponse:
		return msgTypeModifyBearer
	case message.MsgTypeCreateBearerRequest, message.MsgTypeCreateBearerResponse:
		return msgTypeCreateBearer
	case message.MsgTypeUpdateBearerRequest, message.MsgTypeUpdateBearerResponse:
		return msgTypeUpdateBearer
	case message.MsgTypeDeleteBearerRequest, message.MsgTypeDeleteBearerResponse:
		return msgTypeDeleteBearer
	case message.MsgTypeDownlinkDataNotification, message.MsgTypeDownlinkDataNotificationAcknowledge:
		return msgTypeDownlinkDataNotification
	default:
		return causeString(t, true) // reuse the numeric formatter
	}
}

// responseCause returns the Cause IE value and its metric tag: causeNone when
// the message type has no Cause IE (Echo), causeUnknown when the IE is
// missing or cannot be decoded.
func responseCause(msg message.Message) (uint8, string) {
	causeIE, hasCause := causeIEOf(msg)
	if !hasCause {
		return 0, causeNone
	}
	if causeIE == nil {
		return 0, causeUnknown
	}
	v, err := causeIE.Cause()
	if err != nil {
		return 0, causeUnknown
	}
	return v, causeString(v, true)
}

// causeIEOf returns the Cause IE and whether the message type defines one.
// Kept as a type switch so we do not depend on go-gtp exposing a Cause
// accessor method.
func causeIEOf(msg message.Message) (*ie.IE, bool) {
	switch m := msg.(type) {
	case *message.ChangeNotificationResponse:
		return m.Cause, true
	case *message.ContextAcknowledge:
		return m.Cause, true
	case *message.ContextResponse:
		return m.Cause, true
	case *message.CreateBearerResponse:
		return m.Cause, true
	case *message.CreateSessionResponse:
		return m.Cause, true
	case *message.DeleteBearerFailureIndication:
		return m.Cause, true
	case *message.DeleteBearerResponse:
		return m.Cause, true
	case *message.DeletePDNConnectionSetResponse:
		return m.Cause, true
	case *message.DeleteSessionResponse:
		return m.Cause, true
	case *message.DetachAcknowledge:
		return m.Cause, true
	case *message.DownlinkDataNotificationAcknowledge:
		return m.Cause, true
	case *message.DownlinkDataNotificationFailureIndication:
		return m.Cause, true
	case *message.ModifyAccessBearersResponse:
		return m.Cause, true
	case *message.ModifyBearerFailureIndication:
		return m.Cause, true
	case *message.ModifyBearerResponse:
		return m.Cause, true
	case *message.PGWRestartNotificationAcknowledge:
		return m.Cause, true
	case *message.ReleaseAccessBearersResponse:
		return m.Cause, true
	case *message.ResumeAcknowledge:
		return m.Cause, true
	case *message.SuspendAcknowledge:
		return m.Cause, true
	case *message.UpdateBearerResponse:
		return m.Cause, true
	case *message.UpdatePDNConnectionSetResponse:
		return m.Cause, true
	case *message.Generic:
		for _, i := range m.IEs {
			if i != nil && i.Type == ie.Cause {
				return i, true
			}
		}
	}
	return nil, false
}

// TryCreateSessionRaw sends a Create Session Request built from a
// user-supplied list of IEs (typically composed via gtpv2.ie.*) and waits for
// the paired response. imsi is used to attach the peer's F-TEID to the local
// session when one was registered; the response is awaited either way.
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
	_, err := GetMessage[*message.EchoResponse](ctx, c.box, message.MsgTypeEchoResponse, seq)
	c.recordRecvOutcome(msgTypeEcho, 0, err)
	if err != nil {
		return false, err
	}
	return true, nil
}

// CheckRecvCreateSessionResponse waits for the response first and only then
// looks the session up by IMSI: requests that omit the Sender F-TEID have no
// registered session, but their rejection Cause must still be reported.
func (c *K6GTPv2Client) CheckRecvCreateSessionResponse(seq uint32, imsi string) (EnumIFCause, error) {
	causeRes := EnumIFCause(0)
	ctx, cancel := c.recvCtx()
	defer cancel()
	res, err := GetMessage[*message.CreateSessionResponse](ctx, c.box, message.MsgTypeCreateSessionResponse, seq)
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
	if c.Conn == nil {
		return causeRes, nil
	}
	sess, err := c.Conn.GetSessionByIMSI(imsi)
	if err != nil {
		return causeRes, nil
	}
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
	res, err := GetMessage[*message.DeleteSessionResponse](ctx, c.box, message.MsgTypeDeleteSessionResponse, seq)
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
	res, err := GetMessage[*message.ModifyBearerResponse](ctx, c.box, message.MsgTypeModifyBearerResponse, seq)
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

// Close releases the socket and fails every pending wait. It returns only
// after the receive loop has exited, so an immediate Connect can rebind the
// same source address.
func (c *K6GTPv2Client) Close() error {
	return c.shutdown()
}

func (c *K6GTPv2Client) shutdown() error {
	if c.box != nil {
		c.box.close()
	}
	if c.connCancel != nil {
		c.connCancel()
		c.connCancel = nil
	}
	var err error
	if c.Conn != nil {
		err = c.Conn.Close()
		c.Conn = nil
	}
	if c.serveDone != nil {
		<-c.serveDone
		c.serveDone = nil
	}
	return err
}

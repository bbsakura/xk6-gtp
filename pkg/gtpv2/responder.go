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
	"github.com/wmnsk/go-gtp/gtpv2/message"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
)

// K6GTPv2Responder is a lightweight GTPv2-C listener that captures incoming
// requests into an internal queue and lets JS drain them synchronously via
// NextRequest / RespondTo. It is the minimum surface needed to reproduce a
// peer that must accept unsolicited requests (Echo Request from another node,
// Create/Update/Delete Bearer Request from a PGW under test, ...) without
// needing an event-loop callback bridge.
type K6GTPv2Responder struct {
	rm       *RootModule
	vu       modules.VU
	conn     *gtpv2.Conn
	requests chan *IncomingRequest

	serveCtx    context.Context
	serveCancel context.CancelFunc
	serveDone   chan struct{}
	closeOnce   sync.Once
}

// IncomingRequest is the JS-facing view of a captured GTPv2 message. The
// unexported fields carry the pointers RespondTo needs to correlate the
// response with the original request.
type IncomingRequest struct {
	Ok           bool    `json:"ok"`
	MessageType  uint8   `json:"message_type"`
	Sequence     uint32  `json:"sequence"`
	TEID         uint32  `json:"teid"`
	Sender       string  `json:"sender"`
	ReceivedAtMs float64 `json:"received_at_ms"`
	Timeout      bool    `json:"timeout"`
	Error        string  `json:"error"`

	msg        message.Message
	senderAddr net.Addr
}

// ResponderOptions is the JS-facing options struct passed to the Responder
// constructor.
type ResponderOptions struct {
	// Listen is the local UDP address to bind (host:port).
	Listen string `json:"listen"`
	// IfTypeName identifies the local interface type registered on the
	// connection, matching the k6/x/gtpv2 client's ConnectionOptions.
	IfTypeName string `json:"if_type_name"`
	// Count is the GTPv2 Restart Counter advertised in Recovery IEs on the
	// underlying connection.
	Count int `json:"count"`
	// HandleMsgTypes lists the numeric GTPv2 message types (see 3GPP TS
	// 29.274 clause 6.1) that should be captured into the request queue.
	// An empty list captures Echo Request only, which is enough for the
	// smoke-test / node-monitoring use case.
	HandleMsgTypes []uint8 `json:"handle_msg_types"`
	// QueueSize bounds the internal request buffer. Defaults to 32; incoming
	// requests are dropped (with the error metric ticking) once full.
	QueueSize int `json:"queue_size"`
}

// NewK6GTPv2Responder is the JS constructor for the responder. Arguments are
// decoded from a ResponderOptions-shaped object.
func (mi *ModuleInstance) NewK6GTPv2Responder(call sobek.ConstructorCall) *sobek.Object {
	rt := mi.vu.Runtime()
	if len(call.Arguments) == 0 || sobek.IsUndefined(call.Arguments[0]) {
		common.Throw(rt, errors.New("gtpv2: Responder constructor requires an options object"))
	}
	var opts ResponderOptions
	if err := rt.ExportTo(call.Arguments[0], &opts); err != nil {
		common.Throw(rt, fmt.Errorf("gtpv2: decode ResponderOptions: %w", err))
	}

	r, err := mi.newResponder(opts)
	if err != nil {
		common.Throw(rt, err)
	}
	return rt.ToValue(r).ToObject(rt)
}

func (mi *ModuleInstance) newResponder(opts ResponderOptions) (*K6GTPv2Responder, error) {
	if mi.vu == nil || mi.vu.State() == nil {
		return nil, errRunOnly
	}

	laddr, err := net.ResolveUDPAddr("udp", opts.Listen)
	if err != nil {
		return nil, fmt.Errorf("resolve listen addr %q: %w", opts.Listen, err)
	}

	iftype := IFTypeS11MMEGTPC
	if opts.IfTypeName != "" {
		iftype, err = EnumIFTypeString(opts.IfTypeName)
		if err != nil {
			return nil, fmt.Errorf("invalid IfTypeName %q: %w", opts.IfTypeName, err)
		}
	}
	if opts.Count < 0 || opts.Count > 255 {
		return nil, fmt.Errorf("count %d out of range [0, 255]", opts.Count)
	}

	msgTypes := opts.HandleMsgTypes
	if len(msgTypes) == 0 {
		msgTypes = []uint8{message.MsgTypeEchoRequest}
	}
	queueSize := opts.QueueSize
	if queueSize <= 0 {
		queueSize = 32
	}

	// #nosec G115 -- opts.Count is bounded to [0, 255] by the check above.
	conn := gtpv2.NewConn(laddr, uint8(iftype), uint8(opts.Count))

	r := &K6GTPv2Responder{
		rm:        mi.rm,
		vu:        mi.vu,
		conn:      conn,
		requests:  make(chan *IncomingRequest, queueSize),
		serveDone: make(chan struct{}),
	}

	// Register capture handlers before serving so an early packet cannot slip
	// through unhandled.
	for _, mt := range msgTypes {
		conn.AddHandler(mt, r.captureHandler())
	}

	// Bind the listener's lifetime to a context we can cancel from Close.
	// Not tied to vu.Context() so the responder can span iteration boundaries
	// if the script chooses (typical usage: create once in setup, use across
	// iterations, close in teardown).
	r.serveCtx, r.serveCancel = context.WithCancel(context.Background())

	// Bind the socket synchronously so LocalAddr is populated before we
	// return. Only the blocking Serve read loop runs in the background.
	if err := conn.Listen(r.serveCtx); err != nil {
		return nil, fmt.Errorf("gtpv2 listen %s: %w", laddr, err)
	}
	go func() {
		defer close(r.serveDone)
		_ = conn.Serve(r.serveCtx)
	}()

	return r, nil
}

// captureHandler returns a go-gtp handler that pushes the incoming request
// into the responder queue. Full queue drops the message rather than blocking
// the receive goroutine.
func (r *K6GTPv2Responder) captureHandler() func(c *gtpv2.Conn, senderAddr net.Addr, msg message.Message) error {
	return func(_ *gtpv2.Conn, senderAddr net.Addr, msg message.Message) error {
		req := &IncomingRequest{
			Ok:           true,
			MessageType:  msg.MessageType(),
			Sequence:     msg.Sequence(),
			TEID:         msg.TEID(),
			Sender:       senderAddr.String(),
			ReceivedAtMs: elapsedMs(time.Since(time.Unix(0, 0))),
			msg:          msg,
			senderAddr:   senderAddr,
		}
		select {
		case r.requests <- req:
		default:
			// Queue full — drop; the script's polling cadence controls
			// backpressure.
		}
		return nil
	}
}

// NextRequest returns the next captured request, waiting up to timeoutMs
// milliseconds. When the deadline expires (or the responder was closed) the
// returned struct has Ok=false and Timeout=true.
func (r *K6GTPv2Responder) NextRequest(timeoutMs int64) *IncomingRequest {
	if timeoutMs <= 0 {
		timeoutMs = 3000
	}
	select {
	case req, ok := <-r.requests:
		if !ok {
			return &IncomingRequest{Ok: false, Error: "responder closed"}
		}
		return req
	case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
		return &IncomingRequest{Ok: false, Timeout: true}
	case <-r.serveCtx.Done():
		return &IncomingRequest{Ok: false, Error: "responder closed"}
	}
}

// RespondTo sends a GTPv2 response back to the sender of req. Sequence and
// TEID matching are handled by the underlying gtpv2.Conn.RespondTo.
func (r *K6GTPv2Responder) RespondTo(req *IncomingRequest, resp message.Message) error {
	if req == nil || req.senderAddr == nil || req.msg == nil {
		return errors.New("gtpv2: RespondTo requires a request captured by NextRequest")
	}
	return r.conn.RespondTo(req.senderAddr, req.msg, resp)
}

// Close stops the listener and unblocks any pending NextRequest.
func (r *K6GTPv2Responder) Close() error {
	var err error
	r.closeOnce.Do(func() {
		r.serveCancel()
		err = r.conn.Close()
		<-r.serveDone
	})
	return err
}

// LocalAddr returns the address the responder is listening on. Useful when
// listen was given with port 0 and the OS assigned one.
func (r *K6GTPv2Responder) LocalAddr() string {
	if r.conn == nil || r.conn.LocalAddr() == nil {
		return ""
	}
	return r.conn.LocalAddr().String()
}

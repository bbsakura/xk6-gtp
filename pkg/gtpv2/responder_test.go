package gtpv2

import (
	"net"
	"testing"
	"time"

	"github.com/wmnsk/go-gtp/gtpv2/ie"
	"github.com/wmnsk/go-gtp/gtpv2/message"
	"go.k6.io/k6/js/modulestest"
	"go.k6.io/k6/lib"
)

// newTestResponder wires a Responder against a modulestest VU. The VU is put
// into a synthetic run state so requireRunState-guarded call sites can proceed.
func newTestResponder(t *testing.T, opts ResponderOptions) (*K6GTPv2Responder, *ModuleInstance) {
	t.Helper()
	rt := modulestest.NewRuntime(t)
	// Register metrics against the init-context Registry, then move to run
	// context so requireRunState-guarded call sites can proceed.
	rm := New()
	mi := rm.NewModuleInstance(rt.VU).(*ModuleInstance)
	rt.MoveToVUContext(&lib.State{})

	r, err := mi.newResponder(opts)
	if err != nil {
		t.Fatalf("newResponder: %v", err)
	}
	t.Cleanup(func() {
		_ = r.Close()
	})
	return r, mi
}

// TestResponder_EchoRoundtrip runs a Responder on a random port and sends a
// hand-marshaled Echo Request from a bare UDP socket to prove the receive
// path, the queue, and RespondTo all wire up correctly. Uses a raw socket
// (not gtpv2.Dial) so the client's default Echo handshake does not race with
// our capture handler.
func TestResponder_EchoRoundtrip(t *testing.T) {
	r, _ := newTestResponder(t, ResponderOptions{
		Listen:     "127.0.0.1:0",
		IfTypeName: "IFTypeS5S8PGWGTPC",
	})
	waitFor(t, 500*time.Millisecond, func() bool { return r.LocalAddr() != "" })

	raddr, err := net.ResolveUDPAddr("udp", r.LocalAddr())
	if err != nil {
		t.Fatalf("resolve responder addr: %v", err)
	}
	client, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		t.Fatalf("dial udp: %v", err)
	}
	defer client.Close()

	echoReq := message.NewEchoRequest(42, ie.NewRecovery(1))
	echoReq.SetSequenceNumber(42)
	payload, err := message.Marshal(echoReq)
	if err != nil {
		t.Fatalf("marshal echo request: %v", err)
	}
	if _, err := client.Write(payload); err != nil {
		t.Fatalf("write echo request: %v", err)
	}

	req := r.NextRequest(1000)
	if !req.Ok {
		t.Fatalf("NextRequest: ok=false timeout=%v error=%q", req.Timeout, req.Error)
	}
	if req.MessageType != message.MsgTypeEchoRequest {
		t.Fatalf("unexpected message type: %d", req.MessageType)
	}
	if req.Sequence != 42 {
		t.Fatalf("sequence mismatch: got %d, want 42", req.Sequence)
	}

	resp := message.NewEchoResponse(0, ie.NewRecovery(1))
	if err := r.RespondTo(req, resp); err != nil {
		t.Fatalf("RespondTo: %v", err)
	}

	// Confirm a response was written back to our socket.
	if err := client.SetReadDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	buf := make([]byte, 1500)
	n, err := client.Read(buf)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	got, err := message.Parse(buf[:n])
	if err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if got.MessageType() != message.MsgTypeEchoResponse {
		t.Fatalf("unexpected response type: %d", got.MessageType())
	}
	if got.Sequence() != 42 {
		t.Fatalf("response sequence mismatch: got %d, want 42", got.Sequence())
	}
}

func TestResponder_NextRequestTimeout(t *testing.T) {
	r, _ := newTestResponder(t, ResponderOptions{
		Listen:     "127.0.0.1:0",
		IfTypeName: "IFTypeS5S8PGWGTPC",
	})
	waitFor(t, 500*time.Millisecond, func() bool { return r.LocalAddr() != "" })

	start := time.Now()
	req := r.NextRequest(50)
	elapsed := time.Since(start)

	if req.Ok || !req.Timeout {
		t.Fatalf("expected timeout; got %+v", req)
	}
	if elapsed >= 500*time.Millisecond {
		t.Fatalf("NextRequest waited %s past its 50ms budget", elapsed)
	}
}

func TestResponder_CloseReleasesWaiter(t *testing.T) {
	r, _ := newTestResponder(t, ResponderOptions{
		Listen:     "127.0.0.1:0",
		IfTypeName: "IFTypeS5S8PGWGTPC",
	})
	waitFor(t, 500*time.Millisecond, func() bool { return r.LocalAddr() != "" })

	done := make(chan *IncomingRequest, 1)
	go func() {
		done <- r.NextRequest(5000)
	}()

	// Give the goroutine time to block.
	time.Sleep(20 * time.Millisecond)
	if err := r.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	select {
	case req := <-done:
		if req.Ok || req.Error == "" {
			t.Fatalf("expected closed error; got %+v", req)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("NextRequest did not return after Close")
	}
}

func waitFor(t *testing.T, budget time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(budget)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("condition not met within %s", budget)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

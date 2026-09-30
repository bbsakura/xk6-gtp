package refpgw

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/ie"
	"github.com/wmnsk/go-gtp/gtpv2/message"
)

// syncBuffer is a concurrency-safe bytes.Buffer for capturing slog output
// from handler goroutines and reading it back on the test goroutine.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func (b *syncBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]byte, b.buf.Len())
	copy(out, b.buf.Bytes())
	return out
}

// TestHandlers_EchoRoundTrip verifies the extracted Echo handler responds
// with a properly formed Echo Response and emits a structured log record.
func TestHandlers_EchoRoundTrip(t *testing.T) {
	var buf syncBuffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	srvAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}
	srvConn := gtpv2.NewConn(srvAddr, gtpv2.IFTypeS5S8PGWGTPC, 0)
	handlers := New(Config{Logger: logger})
	handlers.AddTo(srvConn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := srvConn.Listen(ctx); err != nil {
		t.Fatalf("server listen: %v", err)
	}
	go func() { _ = srvConn.Serve(ctx) }()
	defer srvConn.Close()

	waitUntil(t, 500*time.Millisecond, func() bool { return srvConn.LocalAddr() != nil })

	raddr := srvConn.LocalAddr().(*net.UDPAddr)
	client, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()

	req := message.NewEchoRequest(7, ie.NewRecovery(0))
	req.SetSequenceNumber(7)
	payload, err := message.Marshal(req)
	if err != nil {
		t.Fatalf("marshal echo request: %v", err)
	}
	if _, err := client.Write(payload); err != nil {
		t.Fatalf("write: %v", err)
	}

	_ = client.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	rbuf := make([]byte, 1500)
	n, err := client.Read(rbuf)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	resp, err := message.Parse(rbuf[:n])
	if err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if resp.MessageType() != message.MsgTypeEchoResponse {
		t.Fatalf("unexpected type: %d", resp.MessageType())
	}
	if resp.Sequence() != 7 {
		t.Fatalf("sequence: got %d want 7", resp.Sequence())
	}

	// Give the handler goroutine a beat to flush the log line.
	waitUntil(t, 200*time.Millisecond, func() bool { return len(buf.Bytes()) > 0 })

	rec := decodeFirstJSON(t, buf.Bytes())
	if rec["msg"] != "received echo request" {
		t.Fatalf("unexpected log msg: %v", rec["msg"])
	}
	if rec["seq"] != float64(7) {
		t.Fatalf("seq attr missing/mismatch: %+v", rec)
	}
	if rec["peer"] == nil {
		t.Fatalf("peer attr missing: %+v", rec)
	}
}

// TestHandlers_DeleteUnknownSession asserts that the Delete Session handler
// returns Cause=IMSIIMEINotKnown and logs the rejection.
func TestHandlers_DeleteUnknownSession(t *testing.T) {
	var buf syncBuffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	srvAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}
	srvConn := gtpv2.NewConn(srvAddr, gtpv2.IFTypeS5S8PGWGTPC, 0)
	// Skip go-gtp's request-validation so the handler's "unknown session"
	// branch is what runs, not the pre-handler reject.
	srvConn.DisableValidation()
	handlers := New(Config{Logger: logger})
	handlers.AddTo(srvConn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := srvConn.Listen(ctx); err != nil {
		t.Fatalf("server listen: %v", err)
	}
	go func() { _ = srvConn.Serve(ctx) }()
	defer srvConn.Close()

	waitUntil(t, 500*time.Millisecond, func() bool { return srvConn.LocalAddr() != nil })

	raddr := srvConn.LocalAddr().(*net.UDPAddr)
	client, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()

	dsr := message.NewDeleteSessionRequest(0xdead, 9, ie.NewEPSBearerID(5))
	dsr.SetSequenceNumber(9)
	payload, err := message.Marshal(dsr)
	if err != nil {
		t.Fatalf("marshal dsr: %v", err)
	}
	if _, err := client.Write(payload); err != nil {
		t.Fatalf("write: %v", err)
	}

	_ = client.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	rbuf := make([]byte, 1500)
	n, err := client.Read(rbuf)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	resp, err := message.Parse(rbuf[:n])
	if err != nil {
		t.Fatalf("parse response: %v", err)
	}
	got, ok := resp.(*message.DeleteSessionResponse)
	if !ok {
		t.Fatalf("unexpected response type %T", resp)
	}
	if got.Cause == nil {
		t.Fatal("cause IE missing")
	}
	code, err := got.Cause.Cause()
	if err != nil {
		t.Fatalf("cause: %v", err)
	}
	if code != gtpv2.CauseIMSIIMEINotKnown {
		t.Fatalf("cause: got %d want %d", code, gtpv2.CauseIMSIIMEINotKnown)
	}

	waitUntil(t, 200*time.Millisecond, func() bool {
		return strings.Contains(buf.String(), "unknown session")
	})
}

func decodeFirstJSON(t *testing.T, raw []byte) map[string]interface{} {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(raw))
	var rec map[string]interface{}
	if err := dec.Decode(&rec); err != nil {
		t.Fatalf("decode log json: %v (buf=%q)", err, string(raw))
	}
	return rec
}

func waitUntil(t *testing.T, budget time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(budget)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("condition not met within %s", budget)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

package gtpv2

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wmnsk/go-gtp/gtpv2/message"
	"go.k6.io/k6/js/modulestest"
)

// TestWaitForMessage_MultiplexedByKey verifies that two concurrent waiters
// on distinct (msgType, seq) keys resolve to their respective messages,
// mirroring the correlation the async Promise API relies on.
func TestWaitForMessage_MultiplexedByKey(t *testing.T) {
	sessions := &sync.Map{}
	msgA := message.NewEchoResponse(1)
	msgB := message.NewEchoResponse(2)

	ready := make(chan struct{}, 2)
	gotA := make(chan message.Message, 1)
	gotB := make(chan message.Message, 1)
	go func() {
		ready <- struct{}{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		m, err := waitForMessage(ctx, sessions, message.MsgTypeEchoResponse, 1)
		if err != nil {
			t.Errorf("waitA: %v", err)
			return
		}
		gotA <- m
	}()
	go func() {
		ready <- struct{}{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		m, err := waitForMessage(ctx, sessions, message.MsgTypeEchoResponse, 2)
		if err != nil {
			t.Errorf("waitB: %v", err)
			return
		}
		gotB <- m
	}()

	<-ready
	<-ready
	time.Sleep(10 * time.Millisecond)

	handler := storeMessageHandler(sessions, message.MsgTypeEchoResponse)
	// Deliver out of order (B first, then A) to prove correlation is by key.
	if err := handler(nil, nil, msgB); err != nil {
		t.Fatalf("handlerB: %v", err)
	}
	if err := handler(nil, nil, msgA); err != nil {
		t.Fatalf("handlerA: %v", err)
	}

	select {
	case m := <-gotA:
		if m != msgA {
			t.Fatalf("waiter A got %v, want %v", m, msgA)
		}
	case <-time.After(time.Second):
		t.Fatal("waiter A never resolved")
	}
	select {
	case m := <-gotB:
		if m != msgB {
			t.Fatalf("waiter B got %v, want %v", m, msgB)
		}
	case <-time.After(time.Second):
		t.Fatal("waiter B never resolved")
	}
}

// TestWaitForMessage_HandlerBeforeWaiter covers the fast-path where the
// receive handler stores the response before the caller starts waiting.
func TestWaitForMessage_HandlerBeforeWaiter(t *testing.T) {
	sessions := &sync.Map{}
	msg := message.NewEchoResponse(0)
	h := storeMessageHandler(sessions, message.MsgTypeEchoResponse)
	if err := h(nil, nil, msg); err != nil {
		t.Fatalf("handler: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	got, err := waitForMessage(ctx, sessions, message.MsgTypeEchoResponse, msg.Sequence())
	if err != nil {
		t.Fatalf("waitForMessage: %v", err)
	}
	if got != msg {
		t.Fatalf("got %v, want %v", got, msg)
	}
	if _, ok := sessions.Load(sessionKey{MessageType: message.MsgTypeEchoResponse, Sequence: msg.Sequence()}); ok {
		t.Fatal("expected pending entry to be cleared")
	}
}

// TestWaitForMessage_WaiterBeforeHandler covers the rendezvous case where the
// caller is already waiting when the handler delivers.
func TestWaitForMessage_WaiterBeforeHandler(t *testing.T) {
	sessions := &sync.Map{}
	msg := message.NewEchoResponse(0)

	ready := make(chan struct{})
	got := make(chan message.Message, 1)
	go func() {
		close(ready)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		m, err := waitForMessage(ctx, sessions, message.MsgTypeEchoResponse, msg.Sequence())
		if err != nil {
			t.Errorf("waitForMessage: %v", err)
			return
		}
		got <- m
	}()

	<-ready
	// Give the waiter a moment to register.
	time.Sleep(10 * time.Millisecond)

	if err := storeMessageHandler(sessions, message.MsgTypeEchoResponse)(nil, nil, msg); err != nil {
		t.Fatalf("handler: %v", err)
	}

	select {
	case m := <-got:
		if m != msg {
			t.Fatalf("got %v, want %v", m, msg)
		}
	case <-time.After(time.Second):
		t.Fatal("waiter did not receive message")
	}
}

// TestWaitForMessage_Timeout ensures cancellation propagates instead of busy
// spinning.
func TestWaitForMessage_Timeout(t *testing.T) {
	sessions := &sync.Map{}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := waitForMessage(ctx, sessions, message.MsgTypeEchoResponse, 42)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error on timeout")
	}
	if elapsed >= 200*time.Millisecond {
		t.Fatalf("waitForMessage took %s, expected ~20ms", elapsed)
	}
}

// TestConnect_InitContextRejected verifies that Connect refuses to open a
// socket when invoked from the init stage (VU state is nil).
func TestConnect_InitContextRejected(t *testing.T) {
	rt := modulestest.NewRuntime(t)
	rm := New()
	mi := rm.NewModuleInstance(rt.VU).(*ModuleInstance)
	cli := &K6GTPv2Client{
		rm:       mi.rm,
		vu:       mi.vu,
		sessions: &sync.Map{},
		timeout:  3,
	}

	_, err := cli.Connect(ConnectionOptions{
		Saddr: "127.0.0.1:0",
		Daddr: "127.0.0.1:0",
	})
	if !errors.Is(err, errRunOnly) {
		t.Fatalf("Connect from init context: got %v, want errRunOnly", err)
	}
}

// TestRootModule_MetricsRegistered verifies that NewModuleInstance registers
// every extension metric against the VU's Registry.
func TestRootModule_MetricsRegistered(t *testing.T) {
	rt := modulestest.NewRuntime(t)
	rm := New()
	_ = rm.NewModuleInstance(rt.VU)

	if rm.metrics == nil {
		t.Fatal("metrics not registered")
	}
	names := []string{
		rm.metrics.reqDuration.Name,
		rm.metrics.reqTotal.Name,
		rm.metrics.respCause.Name,
		rm.metrics.timeoutTotal.Name,
		rm.metrics.sendErrorTotal.Name,
		rm.metrics.connReconnectTotal.Name,
	}
	for _, want := range []string{
		"gtpv2_req_duration",
		"gtpv2_req_total",
		"gtpv2_resp_cause",
		"gtpv2_timeout_total",
		"gtpv2_send_error_total",
		"gtpv2_conn_reconnect_total",
	} {
		found := false
		for _, got := range names {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("metric %q not registered", want)
		}
	}
}

// TestRootModule_OnceAcrossVUs verifies that once.Do runs a single time even
// as multiple VU instances race NewModuleInstance concurrently.
func TestRootModule_OnceAcrossVUs(t *testing.T) {
	rm := New()

	var runs int32
	rm.metricsOnce.Do(func() { atomic.AddInt32(&runs, 1) })

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rt := modulestest.NewRuntime(t)
			_ = rm.NewModuleInstance(rt.VU)
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&runs); got != 1 {
		t.Fatalf("once.Do ran %d times, want 1", got)
	}
}

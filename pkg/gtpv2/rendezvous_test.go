package gtpv2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wmnsk/go-gtp/gtpv2/ie"
	"github.com/wmnsk/go-gtp/gtpv2/message"
)

func waitAsync(box *responseBox, msgType uint8, seq uint32, timeout time.Duration) <-chan error {
	done := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		_, err := box.wait(ctx, msgType, seq)
		done <- err
	}()
	return done
}

func waitPending(t *testing.T, box *responseBox, msgType uint8, seq uint32) {
	t.Helper()
	waitFor(t, time.Second, func() bool {
		box.mu.Lock()
		defer box.mu.Unlock()
		e, ok := box.entries[sessionKey{MessageType: msgType, Sequence: seq}]
		return ok && e.ch != nil
	})
}

func TestResponseBox_SecondAwaitIsRejectedAndFirstStillResolves(t *testing.T) {
	box := newResponseBox()
	first := waitAsync(box, message.MsgTypeEchoResponse, 7, time.Second)
	waitPending(t, box, message.MsgTypeEchoResponse, 7)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := box.wait(ctx, message.MsgTypeEchoResponse, 7); !errors.Is(err, errDuplicateAwait) {
		t.Fatalf("second wait: got %v, want errDuplicateAwait", err)
	}

	box.deliver(message.MsgTypeEchoResponse, message.NewEchoResponse(7))
	if err := <-first; err != nil {
		t.Fatalf("first waiter: %v", err)
	}
}

func TestResponseBox_DuplicateResponseWhileBufferedKeepsFirst(t *testing.T) {
	box := newResponseBox()
	firstMsg := message.NewEchoResponse(3)
	box.deliver(message.MsgTypeEchoResponse, firstMsg)
	box.deliver(message.MsgTypeEchoResponse, message.NewEchoResponse(3))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	got, err := box.wait(ctx, message.MsgTypeEchoResponse, 3)
	if err != nil {
		t.Fatalf("wait: %v", err)
	}
	if got != firstMsg {
		t.Fatalf("got %p, want the first delivered message %p", got, firstMsg)
	}
}

func TestResponseBox_DuplicateResponseAfterWaiterSignalledDoesNotPanic(t *testing.T) {
	box := newResponseBox()
	done := waitAsync(box, message.MsgTypeEchoResponse, 4, time.Second)
	waitPending(t, box, message.MsgTypeEchoResponse, 4)

	box.deliver(message.MsgTypeEchoResponse, message.NewEchoResponse(4))
	box.deliver(message.MsgTypeEchoResponse, message.NewEchoResponse(4))
	if err := <-done; err != nil {
		t.Fatalf("wait: %v", err)
	}
}

func TestResponseBox_CloseFailsPendingWaitImmediately(t *testing.T) {
	box := newResponseBox()
	done := waitAsync(box, message.MsgTypeEchoResponse, 5, 5*time.Second)
	waitPending(t, box, message.MsgTypeEchoResponse, 5)

	box.close()
	select {
	case err := <-done:
		if !errors.Is(err, errConnClosed) {
			t.Fatalf("got %v, want errConnClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("pending wait was not released by close")
	}
}

func TestResponseBox_ClosedBoxIgnoresLateResponses(t *testing.T) {
	old := newResponseBox()
	old.deliver(message.MsgTypeEchoResponse, message.NewEchoResponse(1))
	old.close()
	old.deliver(message.MsgTypeEchoResponse, message.NewEchoResponse(1))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := old.wait(ctx, message.MsgTypeEchoResponse, 1); !errors.Is(err, errConnClosed) {
		t.Fatalf("wait on closed box: got %v, want errConnClosed", err)
	}
}

func TestResponseCause_DistinguishesNoCauseFromMissingCause(t *testing.T) {
	for name, tc := range map[string]struct {
		msg      message.Message
		wantCode uint8
		wantTag  string
	}{
		"echo has no cause":            {message.NewEchoResponse(1), 0, causeNone},
		"create session without cause": {message.NewCreateSessionResponse(0, 1), 0, causeUnknown},
		"create session with cause":    {message.NewCreateSessionResponse(0, 1, ie.NewCause(64, 0, 0, 0, nil)), 64, "64"},
		"downlink data notification ack": {
			message.NewDownlinkDataNotificationAcknowledge(0, 1, ie.NewCause(16, 0, 0, 0, nil)), 16, "16",
		},
	} {
		t.Run(name, func(t *testing.T) {
			code, tag := responseCause(tc.msg)
			if code != tc.wantCode || tag != tc.wantTag {
				t.Fatalf("got (%d, %q), want (%d, %q)", code, tag, tc.wantCode, tc.wantTag)
			}
		})
	}
}

func TestMessageTypeTag_RequestAndResponseShareTag(t *testing.T) {
	for req, resp := range map[uint8]uint8{
		message.MsgTypeEchoRequest:          message.MsgTypeEchoResponse,
		message.MsgTypeCreateSessionRequest: message.MsgTypeCreateSessionResponse,
		message.MsgTypeDeleteSessionRequest: message.MsgTypeDeleteSessionResponse,
		message.MsgTypeModifyBearerRequest:  message.MsgTypeModifyBearerResponse,
		message.MsgTypeCreateBearerRequest:  message.MsgTypeCreateBearerResponse,
	} {
		if messageTypeTag(req) != messageTypeTag(resp) {
			t.Errorf("tag(%d)=%q, tag(%d)=%q", req, messageTypeTag(req), resp, messageTypeTag(resp))
		}
	}
}

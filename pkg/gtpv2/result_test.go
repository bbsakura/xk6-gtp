package gtpv2

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewSendError_Timeout(t *testing.T) {
	r := newSendError(50*time.Millisecond, context.DeadlineExceeded)
	if r.Ok {
		t.Fatal("ok should be false")
	}
	if !r.Timeout {
		t.Fatal("timeout flag not set for DeadlineExceeded")
	}
	if r.ElapsedMs != 50 {
		t.Fatalf("elapsed ms: got %v, want 50", r.ElapsedMs)
	}
	if r.Error == "" {
		t.Fatal("error string missing")
	}
}

func TestNewSendError_Wrapped(t *testing.T) {
	inner := errors.New("dial refused")
	r := newSendError(10*time.Millisecond, inner)
	if r.Timeout {
		t.Fatal("non-timeout error should not set Timeout")
	}
	if r.Error != inner.Error() {
		t.Fatalf("error message: got %q, want %q", r.Error, inner.Error())
	}
}

func TestNewSendOK(t *testing.T) {
	r := newSendOK(42, 16, 12500*time.Microsecond)
	if !r.Ok {
		t.Fatal("ok should be true")
	}
	if r.Sequence != 42 {
		t.Fatalf("sequence: got %d", r.Sequence)
	}
	if r.Cause != 16 {
		t.Fatalf("cause: got %d", r.Cause)
	}
	if r.ElapsedMs != 12.5 {
		t.Fatalf("elapsed_ms: got %v", r.ElapsedMs)
	}
	if r.Timeout {
		t.Fatal("timeout should be false on success")
	}
}

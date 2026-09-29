package gtpv2

import (
	"context"
	"errors"
	"time"
)

// SendResult is the rich return type for the Try* combined send/receive
// helpers. Exposed to JS with underscore-cased field names so tests can
// destructure it directly.
type SendResult struct {
	Ok        bool    `json:"ok"`
	Sequence  uint32  `json:"sequence"`
	Cause     uint8   `json:"cause"`
	ElapsedMs float64 `json:"elapsed_ms"`
	Timeout   bool    `json:"timeout"`
	Error     string  `json:"error"`
}

func newSendError(elapsed time.Duration, err error) *SendResult {
	return &SendResult{
		Ok:        false,
		ElapsedMs: elapsedMs(elapsed),
		Timeout:   errors.Is(err, context.DeadlineExceeded),
		Error:     err.Error(),
	}
}

func newSendOK(seq uint32, cause uint8, elapsed time.Duration) *SendResult {
	return &SendResult{
		Ok:        true,
		Sequence:  seq,
		Cause:     cause,
		ElapsedMs: elapsedMs(elapsed),
	}
}

func elapsedMs(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

// SendHandle is the return type of split-send methods (SendRaw). Pair the
// sequence with AwaitMessage to inspect the response.
type SendHandle struct {
	Ok       bool   `json:"ok"`
	Sequence uint32 `json:"sequence"`
	Error    string `json:"error"`
}

// AwaitResult is the return type of AwaitMessage. Cause is 0 when the awaited
// message type does not carry a Cause IE.
type AwaitResult struct {
	Ok        bool    `json:"ok"`
	Cause     uint8   `json:"cause"`
	ElapsedMs float64 `json:"elapsed_ms"`
	Timeout   bool    `json:"timeout"`
	Error     string  `json:"error"`
}

func newAwaitError(elapsed time.Duration, err error) *AwaitResult {
	return &AwaitResult{
		Ok:        false,
		ElapsedMs: elapsedMs(elapsed),
		Timeout:   errors.Is(err, context.DeadlineExceeded),
		Error:     err.Error(),
	}
}

func newAwaitOK(cause uint8, elapsed time.Duration) *AwaitResult {
	return &AwaitResult{
		Ok:        true,
		Cause:     cause,
		ElapsedMs: elapsedMs(elapsed),
	}
}

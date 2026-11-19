package gtpv2

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"sync"
	"time"

	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/message"
)

var (
	errConnClosed     = errors.New("gtpv2: connection closed while waiting for response")
	errDuplicateAwait = errors.New("gtpv2: another wait for the same message type and sequence is already pending")
)

type sessionKey struct {
	MessageType uint8
	Sequence    uint32
}

// pendingEntry is the rendezvous slot for one (msgType, seq). ch is nil when
// the response arrived before anyone waited for it.
type pendingEntry struct {
	msg       message.Message
	ch        chan struct{}
	delivered bool
}

// responseBox correlates responses with waiters for a single gtpv2.Conn.
// Every Connect creates a new box and Close shuts it, so handlers bound to a
// previous connection can never satisfy waits on the next one.
type responseBox struct {
	mu      sync.Mutex
	entries map[sessionKey]*pendingEntry
	sentAt  map[uint32]time.Time
	done    chan struct{}
	closed  bool
}

func newResponseBox() *responseBox {
	return &responseBox{
		entries: make(map[sessionKey]*pendingEntry),
		sentAt:  make(map[uint32]time.Time),
		done:    make(chan struct{}),
	}
}

// deliver hands msg to the waiter for its key, or buffers it. Duplicates of a
// response that is still buffered or already handed to a waiter are dropped.
func (b *responseBox) deliver(msgType uint8, msg message.Message) {
	key := sessionKey{MessageType: msgType, Sequence: msg.Sequence()}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	e, ok := b.entries[key]
	if !ok {
		b.entries[key] = &pendingEntry{msg: msg}
		return
	}
	if e.ch == nil || e.delivered {
		return
	}
	e.msg = msg
	e.delivered = true
	close(e.ch)
}

func (b *responseBox) wait(ctx context.Context, msgType uint8, seq uint32) (message.Message, error) {
	if b == nil {
		return nil, errNotConnected
	}
	key := sessionKey{MessageType: msgType, Sequence: seq}
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil, errConnClosed
	}
	if e, ok := b.entries[key]; ok {
		if e.ch != nil {
			b.mu.Unlock()
			return nil, errDuplicateAwait
		}
		delete(b.entries, key)
		b.mu.Unlock()
		return e.msg, nil
	}
	e := &pendingEntry{ch: make(chan struct{})}
	b.entries[key] = e
	b.mu.Unlock()

	var err error
	select {
	case <-e.ch:
	case <-b.done:
		err = errConnClosed
	case <-ctx.Done():
		err = ctx.Err()
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.entries[key] == e {
		delete(b.entries, key)
	}
	// A delivery that raced with cancellation still wins.
	if e.delivered {
		return e.msg, nil
	}
	return nil, err
}

func (b *responseBox) markSent(seq uint32, at time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.closed {
		b.sentAt[seq] = at
	}
}

func (b *responseBox) takeSentAt(seq uint32) (time.Time, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	t, ok := b.sentAt[seq]
	delete(b.sentAt, seq)
	return t, ok
}

func (b *responseBox) close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	close(b.done)
	b.entries = nil
	b.sentAt = nil
}

// GetMessage waits for a GTPv2 response matching (msgType, seq) and asserts its
// concrete type.
func GetMessage[PT *T, T any](ctx context.Context, box *responseBox, msgType uint8, seq uint32) (PT, error) {
	msg, err := box.wait(ctx, msgType, seq)
	if err != nil {
		return PT(nil), err
	}
	typed, ok := msg.(PT)
	if !ok {
		return PT(nil), fmt.Errorf("gtpv2: unexpected message type %T for seq=%d", msg, seq)
	}
	return typed, nil
}

func storeMessageHandler(box *responseBox, msgType uint8) gtpv2.HandlerFunc {
	return func(_ *gtpv2.Conn, _ net.Addr, msg message.Message) error {
		box.deliver(msgType, msg)
		return nil
	}
}

// responseMsgTypes are captured from the moment the connection starts serving,
// so a reply cannot arrive before its handler exists. Request types are left
// to go-gtp's defaults (e.g. automatic Echo Response).
var responseMsgTypes = []uint8{
	message.MsgTypeEchoResponse,
	message.MsgTypeAlertMMEAcknowledge,
	message.MsgTypeBearerResourceFailureIndication,
	message.MsgTypeChangeNotificationResponse,
	message.MsgTypeContextAcknowledge,
	message.MsgTypeContextResponse,
	message.MsgTypeCreateBearerResponse,
	message.MsgTypeCreateForwardingTunnelResponse,
	message.MsgTypeCreateIndirectDataForwardingTunnelResponse,
	message.MsgTypeCreateSessionResponse,
	message.MsgTypeDeleteBearerFailureIndication,
	message.MsgTypeDeleteBearerResponse,
	message.MsgTypeDeleteIndirectDataForwardingTunnelResponse,
	message.MsgTypeDeletePDNConnectionSetResponse,
	message.MsgTypeDeleteSessionResponse,
	message.MsgTypeDetachAcknowledge,
	message.MsgTypeDirectTransferResponse,
	message.MsgTypeDownlinkDataNotificationAcknowledge,
	message.MsgTypeDownlinkDataNotificationFailureIndication,
	message.MsgTypeForwardAccessContextAcknowledge,
	message.MsgTypeForwardRelocationCompleteAcknowledge,
	message.MsgTypeForwardRelocationResponse,
	message.MsgTypeIdentificationResponse,
	message.MsgTypeMBMSSessionStartResponse,
	message.MsgTypeMBMSSessionStopResponse,
	message.MsgTypeMBMSSessionUpdateResponse,
	message.MsgTypeModifyAccessBearersResponse,
	message.MsgTypeModifyBearerFailureIndication,
	message.MsgTypeModifyBearerResponse,
	message.MsgTypeNotificationResponse,
	message.MsgTypePGWDownlinkTriggeringAcknowledge,
	message.MsgTypePGWRestartNotificationAcknowledge,
	message.MsgTypeReleaseAccessBearersResponse,
	message.MsgTypeRelocationCancelResponse,
	message.MsgTypeRemoteUEReportAcknowledge,
	message.MsgTypeResumeAcknowledge,
	message.MsgTypeSRVCCCsToPsCancelAcknowledge,
	message.MsgTypeSRVCCCsToPsCompleteAcknowledge,
	message.MsgTypeSRVCCCsToPsResponse,
	message.MsgTypeSRVCCPsToCsCancelAcknowledge,
	message.MsgTypeSRVCCPsToCsCompleteAcknowledge,
	message.MsgTypeSRVCCPsToCsResponse,
	message.MsgTypeSuspendAcknowledge,
	message.MsgTypeUEActivityAcknowledge,
	message.MsgTypeUERegistrationQueryResponse,
	message.MsgTypeUpdateBearerResponse,
	message.MsgTypeUpdatePDNConnectionSetResponse,
}

func isPreregistered(msgType uint8) bool {
	return slices.Contains(responseMsgTypes, msgType)
}

func setHandlers(conn *gtpv2.Conn, box *responseBox) {
	for _, mt := range responseMsgTypes {
		conn.AddHandler(mt, storeMessageHandler(box, mt))
	}
}

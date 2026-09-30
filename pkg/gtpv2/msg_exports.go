package gtpv2

import (
	"github.com/wmnsk/go-gtp/gtpv2/ie"
	"github.com/wmnsk/go-gtp/gtpv2/message"
)

// msgExports returns the GTPv2 message constructors exposed to JS as
// `gtpv2.msg.*`. These pair with `gtpv2.ie.*` and `client.sendRaw` to let
// scripts assemble any GTPv2 message end-to-end from JavaScript, so
// abnormal-sequence testing (unexpected TEID, non-standard IE mix, PGW-
// initiated commands, ...) does not need a Go change.
//
// The go-gtp constructors accept a variadic `...*ie.IE`; sobek does not
// auto-spread a JS array into Go variadic arguments, so every entry here is
// wrapped to take a `[]*ie.IE` slice.
func msgExports() map[string]interface{} {
	return map[string]interface{}{
		// SGW/MME → PGW initiator side
		"newEchoRequest":               wrapNoTEID(message.NewEchoRequest),
		"newCreateSessionRequest":      wrapTEID(message.NewCreateSessionRequest),
		"newDeleteSessionRequest":      wrapTEID(message.NewDeleteSessionRequest),
		"newModifyBearerRequest":       wrapTEID(message.NewModifyBearerRequest),
		"newModifyBearerCommand":       wrapTEID(message.NewModifyBearerCommand),
		"newDeleteBearerCommand":       wrapTEID(message.NewDeleteBearerCommand),
		"newChangeNotificationRequest": wrapTEID(message.NewChangeNotificationRequest),

		// PGW → SGW/MME initiator side (typically used by responders, kept
		// here so raw scripts can also drive them).
		"newCreateBearerRequest":      wrapTEID(message.NewCreateBearerRequest),
		"newUpdateBearerRequest":      wrapTEID(message.NewUpdateBearerRequest),
		"newDeleteBearerRequest":      wrapTEID(message.NewDeleteBearerRequest),
		"newDownlinkDataNotification": wrapTEID(message.NewDownlinkDataNotification),

		// Response side
		"newEchoResponse":                        wrapNoTEID(message.NewEchoResponse),
		"newCreateSessionResponse":               wrapTEID(message.NewCreateSessionResponse),
		"newDeleteSessionResponse":               wrapTEID(message.NewDeleteSessionResponse),
		"newModifyBearerResponse":                wrapTEID(message.NewModifyBearerResponse),
		"newCreateBearerResponse":                wrapTEID(message.NewCreateBearerResponse),
		"newUpdateBearerResponse":                wrapTEID(message.NewUpdateBearerResponse),
		"newDeleteBearerResponse":                wrapTEID(message.NewDeleteBearerResponse),
		"newDownlinkDataNotificationAcknowledge": wrapTEID(message.NewDownlinkDataNotificationAcknowledge),

		// Generic escape hatch for message types not enumerated above.
		"newGeneric":            newGenericJS,
		"newGenericWithoutTEID": newGenericWithoutTEIDJS,
	}
}

// wrapTEID adapts a go-gtp `func(teid, seq uint32, ies ...*ie.IE) *T`
// constructor to a JS-friendly `func(teid, seq uint32, ies []*ie.IE) *T`.
// The returned message is upcast to message.Message so the map value type
// stays uniform.
func wrapTEID[T message.Message](fn func(teid, seq uint32, ies ...*ie.IE) T) func(teid, seq uint32, ies []*ie.IE) message.Message {
	return func(teid, seq uint32, ies []*ie.IE) message.Message {
		return fn(teid, seq, ies...)
	}
}

// wrapNoTEID adapts a `func(seq uint32, ies ...*ie.IE) *T` constructor.
func wrapNoTEID[T message.Message](fn func(seq uint32, ies ...*ie.IE) T) func(seq uint32, ies []*ie.IE) message.Message {
	return func(seq uint32, ies []*ie.IE) message.Message {
		return fn(seq, ies...)
	}
}

func newGenericJS(msgType uint8, teid, seq uint32, ies []*ie.IE) message.Message {
	return message.NewGeneric(msgType, teid, seq, ies...)
}

func newGenericWithoutTEIDJS(msgType uint8, teid, seq uint32, ies []*ie.IE) message.Message {
	return message.NewGenericWithoutTEID(msgType, teid, seq, ies...)
}

package gtpv2

import (
	"context"
	"strconv"
	"time"

	"go.k6.io/k6/lib"
	"go.k6.io/k6/metrics"
)

const (
	tagMsgType = "msg_type"
	tagCause   = "cause"
	tagError   = "error"
	tagPeer    = "peer"

	msgTypeEcho          = "echo"
	msgTypeCreateSession = "create_session"
	msgTypeDeleteSession = "delete_session"
	msgTypeModifyBearer  = "modify_bearer"

	// causeNone marks responses whose message type does not carry a Cause IE
	// (Echo). Kept distinct from "unknown" (Cause absent or unparseable) so
	// dashboards can spot decoding failures without hiding successful echoes.
	causeNone    = "none"
	causeUnknown = "unknown"
)

// gtpMetrics is the k6 metric set exposed by this extension. Registered
// exactly once per k6 process via RootModule.NewModuleInstance.
type gtpMetrics struct {
	reqDuration        *metrics.Metric
	reqTotal           *metrics.Metric
	respCause          *metrics.Metric
	timeoutTotal       *metrics.Metric
	sendErrorTotal     *metrics.Metric
	connReconnectTotal *metrics.Metric
}

func registerMetrics(reg *metrics.Registry) *gtpMetrics {
	return &gtpMetrics{
		reqDuration:        reg.MustNewMetric("gtpv2_req_duration", metrics.Trend, metrics.Time),
		reqTotal:           reg.MustNewMetric("gtpv2_req_total", metrics.Counter),
		respCause:          reg.MustNewMetric("gtpv2_resp_cause", metrics.Counter),
		timeoutTotal:       reg.MustNewMetric("gtpv2_timeout_total", metrics.Counter),
		sendErrorTotal:     reg.MustNewMetric("gtpv2_send_error_total", metrics.Counter),
		connReconnectTotal: reg.MustNewMetric("gtpv2_conn_reconnect_total", metrics.Counter),
	}
}

// pushRequest emits req_total tagged by message type when a request has been
// dispatched to the wire.
func (m *gtpMetrics) pushRequest(ctx context.Context, state *lib.State, msgType string) {
	if m == nil || state == nil {
		return
	}
	tags := state.Tags.GetCurrentValues().Tags.With(tagMsgType, msgType)
	metrics.PushIfNotDone(ctx, state.Samples, metrics.Sample{
		TimeSeries: metrics.TimeSeries{Metric: m.reqTotal, Tags: tags},
		Time:       time.Now(),
		Value:      1,
	})
}

// pushResponse emits resp_cause for a delivered response. cause is the numeric
// Cause IE value formatted as decimal, or one of the causeNone / causeUnknown
// sentinels.
func (m *gtpMetrics) pushResponse(ctx context.Context, state *lib.State, msgType, cause string) {
	if m == nil || state == nil {
		return
	}
	base := state.Tags.GetCurrentValues().Tags
	tags := base.With(tagMsgType, msgType).With(tagCause, cause)
	metrics.PushIfNotDone(ctx, state.Samples, metrics.Sample{
		TimeSeries: metrics.TimeSeries{Metric: m.respCause, Tags: tags},
		Time:       time.Now(),
		Value:      1,
	})
}

// pushTimeout emits timeout_total tagged by message type when a receive wait
// expired without a matching response.
func (m *gtpMetrics) pushTimeout(ctx context.Context, state *lib.State, msgType string) {
	if m == nil || state == nil {
		return
	}
	tags := state.Tags.GetCurrentValues().Tags.With(tagMsgType, msgType)
	metrics.PushIfNotDone(ctx, state.Samples, metrics.Sample{
		TimeSeries: metrics.TimeSeries{Metric: m.timeoutTotal, Tags: tags},
		Time:       time.Now(),
		Value:      1,
	})
}

// pushSendError emits send_error_total when a request could not be dispatched
// to the wire (address resolution failure, socket closed, etc). Distinct from
// timeout_total, which fires on receive-side deadlines.
func (m *gtpMetrics) pushSendError(ctx context.Context, state *lib.State, msgType string) {
	if m == nil || state == nil {
		return
	}
	tags := state.Tags.GetCurrentValues().Tags.With(tagMsgType, msgType)
	metrics.PushIfNotDone(ctx, state.Samples, metrics.Sample{
		TimeSeries: metrics.TimeSeries{Metric: m.sendErrorTotal, Tags: tags},
		Time:       time.Now(),
		Value:      1,
	})
}

// pushReconnect emits conn_reconnect_total when a Client re-establishes its
// underlying gtpv2.Conn (Connect after Close). The peer address is tagged so
// dashboards can spot flapping on a specific PGW/SGW pair.
func (m *gtpMetrics) pushReconnect(ctx context.Context, state *lib.State, peer string) {
	if m == nil || state == nil {
		return
	}
	tags := state.Tags.GetCurrentValues().Tags.With(tagPeer, peer)
	metrics.PushIfNotDone(ctx, state.Samples, metrics.Sample{
		TimeSeries: metrics.TimeSeries{Metric: m.connReconnectTotal, Tags: tags},
		Time:       time.Now(),
		Value:      1,
	})
}

// pushDuration emits req_duration for a full send/receive round trip. isErr is
// exposed as an "error" tag so success and failure trends can be split.
func (m *gtpMetrics) pushDuration(ctx context.Context, state *lib.State, msgType string, elapsed time.Duration, isErr bool) {
	if m == nil || state == nil {
		return
	}
	tags := state.Tags.GetCurrentValues().Tags.With(tagMsgType, msgType)
	if isErr {
		tags = tags.With(tagError, "true")
	}
	metrics.PushIfNotDone(ctx, state.Samples, metrics.Sample{
		TimeSeries: metrics.TimeSeries{Metric: m.reqDuration, Tags: tags},
		Time:       time.Now(),
		Value:      metrics.D(elapsed),
	})
}

// causeString renders a numeric Cause IE value as a low-cardinality tag. The
// GTPv2 Cause enum is bounded (< 128 documented values) so per-value tags are
// acceptable; per-session identifiers must never land here.
func causeString(code uint8, ok bool) string {
	if !ok {
		return causeUnknown
	}
	return strconv.FormatUint(uint64(code), 10)
}

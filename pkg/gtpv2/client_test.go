package gtpv2

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/bbsakura/xk6-gtp/pkg/refpgw"
	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/ie"
	"github.com/wmnsk/go-gtp/gtpv2/message"
	"go.k6.io/k6/js/modulestest"
	"go.k6.io/k6/lib"
	"go.k6.io/k6/metrics"
)

func startRefPGW(t *testing.T) string {
	t.Helper()
	conn := gtpv2.NewConn(&net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}, gtpv2.IFTypeS5S8PGWGTPC, 0)
	refpgw.New(refpgw.Config{}).AddTo(conn)
	ctx, cancel := context.WithCancel(context.Background())
	if err := conn.Listen(ctx); err != nil {
		cancel()
		t.Fatalf("pgw listen: %v", err)
	}
	served := make(chan struct{})
	go func() {
		defer close(served)
		_ = conn.Serve(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		<-served
	})
	return conn.LocalAddr().String()
}

func freeUDPAddr(t *testing.T) string {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := pc.LocalAddr().String()
	_ = pc.Close()
	return addr
}

// newRunClient returns a Client whose VU is in the run stage, plus the
// channel its samples are pushed to.
func newRunClient(t *testing.T) (*K6GTPv2Client, chan metrics.SampleContainer) {
	t.Helper()
	rt := modulestest.NewRuntime(t)
	rm := New()
	mi := rm.NewModuleInstance(rt.VU).(*ModuleInstance)
	samples := make(chan metrics.SampleContainer, 1000)
	rt.MoveToVUContext(&lib.State{
		Samples: samples,
		Tags:    lib.NewVUStateTags(rt.VU.InitEnvField.Registry.RootTagSet()),
	})
	cli := &K6GTPv2Client{rm: rm, vu: mi.vu, timeout: 1}
	t.Cleanup(func() { _ = cli.Close() })
	return cli, samples
}

func countSamples(samples chan metrics.SampleContainer, metric, msgType string) int {
	n := 0
	for {
		select {
		case sc := <-samples:
			for _, s := range sc.GetSamples() {
				if s.Metric.Name != metric {
					continue
				}
				if v, _ := s.Tags.Get(tagMsgType); msgType == "" || v == msgType {
					n++
				}
			}
		default:
			return n
		}
	}
}

func s5s8Params() S5S8SgwParams {
	return S5S8SgwParams{
		Imsi: "123451234567891", Msisdn: "123451234567891", Mei: "123451234567891",
		Mcc: "123", Mnc: "123", Tac: 1, Eci: 1, Apn: "apn", PdnType: 1, Epsbearerid: 5,
		UplaneIE: TEIDParams{Teid: 1}, Ambrul: 1, Ambrdl: 1,
	}
}

func TestClient_EntryPointsReturnErrorsBeforeConnect(t *testing.T) {
	cli, samples := newRunClient(t)

	if r := cli.TryEcho("127.0.0.1:2123"); r.Ok || r.Error != errNotConnected.Error() {
		t.Errorf("TryEcho: %+v", r)
	}
	if r := cli.TryCreateSessionS5S8("127.0.0.1:2123", s5s8Params()); r.Ok || r.Error != errNotConnected.Error() {
		t.Errorf("TryCreateSessionS5S8: %+v", r)
	}
	if r := cli.TryModifyBearerS5S8("", s5s8Params()); r.Ok || r.Error != errNotConnected.Error() {
		t.Errorf("TryModifyBearerS5S8: %+v", r)
	}
	if r := cli.SendRaw("127.0.0.1:2123", message.NewEchoRequest(0)); r.Ok || r.Error != errNotConnected.Error() {
		t.Errorf("SendRaw: %+v", r)
	}
	if r := cli.SendRaw("127.0.0.1:2123", nil); r.Ok {
		t.Errorf("SendRaw(nil): %+v", r)
	}
	if r := cli.AwaitMessage(message.MsgTypeEchoResponse, 1, 10); r.Ok || r.Error != errNotConnected.Error() {
		t.Errorf("AwaitMessage: %+v", r)
	}
	if _, err := cli.CheckRecvEchoResponse(1); !errors.Is(err, errNotConnected) {
		t.Errorf("CheckRecvEchoResponse: %v", err)
	}

	if got := countSamples(samples, "gtpv2_send_error_total", msgTypeModifyBearer); got != 1 {
		t.Errorf("modify_bearer send errors: got %d, want 1", got)
	}
}

func TestClient_DeleteLookupFailureCountsOneSendError(t *testing.T) {
	pgw := startRefPGW(t)
	cli, samples := newRunClient(t)
	if _, err := cli.Connect(ConnectionOptions{Saddr: "127.0.0.1:0", Daddr: pgw, IfTypeName: "IFTypeS5S8SGWGTPC"}); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	countSamples(samples, "", "") // drain Connect-time samples

	p := s5s8Params()
	p.Imsi = "999999999999999"
	if r := cli.TryDeleteSessionS5S8("", p); r.Ok {
		t.Fatalf("delete of unknown IMSI succeeded: %+v", r)
	}
	if got := countSamples(samples, "gtpv2_send_error_total", msgTypeDeleteSession); got != 1 {
		t.Errorf("delete_session send errors: got %d, want 1", got)
	}
}

func TestClient_ReconnectRebindsSameSourceAddress(t *testing.T) {
	pgw := startRefPGW(t)
	cli, _ := newRunClient(t)
	opts := ConnectionOptions{Saddr: freeUDPAddr(t), Daddr: pgw, IfTypeName: "IFTypeS5S8SGWGTPC"}
	for i := range 20 {
		if _, err := cli.Connect(opts); err != nil {
			t.Fatalf("Connect #%d: %v", i, err)
		}
		if err := cli.Close(); err != nil {
			t.Fatalf("Close #%d: %v", i, err)
		}
	}
}

func TestClient_CloseFailsPendingAwait(t *testing.T) {
	pgw := startRefPGW(t)
	cli, _ := newRunClient(t)
	if _, err := cli.Connect(ConnectionOptions{Saddr: "127.0.0.1:0", Daddr: pgw, IfTypeName: "IFTypeS5S8SGWGTPC"}); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	box := cli.box
	done := make(chan *AwaitResult, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		done <- cli.awaitAndRecord(ctx, box, message.MsgTypeEchoResponse, 999)
	}()
	waitPending(t, box, message.MsgTypeEchoResponse, 999)

	_ = cli.Close()
	select {
	case r := <-done:
		if r.Ok || r.Error != errConnClosed.Error() {
			t.Fatalf("got %+v, want errConnClosed", r)
		}
	case <-time.After(time.Second):
		t.Fatal("await was not released by Close")
	}
}

func TestClient_SendRawCreateSessionRecordsCauseAndDuration(t *testing.T) {
	pgw := startRefPGW(t)
	cli, samples := newRunClient(t)
	if _, err := cli.Connect(ConnectionOptions{Saddr: "127.0.0.1:0", Daddr: pgw, IfTypeName: "IFTypeS5S8SGWGTPC"}); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	p := s5s8Params()
	cteid := ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8SGWGTPC, 0x1234, "127.0.0.1", "")
	uteid := ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8SGWGTPU, 1, "127.0.0.1", "").WithInstance(2)
	msg := message.NewCreateSessionRequest(0, 0, cli.genS5S8SessionIE(p, cteid, uteid, "127.0.0.1")...)

	h := cli.SendRaw(pgw, msg)
	if !h.Ok {
		t.Fatalf("SendRaw: %+v", h)
	}
	r := cli.AwaitMessage(message.MsgTypeCreateSessionResponse, h.Sequence, 1000)
	if !r.Ok || r.Cause != gtpv2.CauseRequestAccepted {
		t.Fatalf("AwaitMessage: %+v", r)
	}
	if got := countSamples(samples, "gtpv2_req_duration", msgTypeCreateSession); got != 1 {
		t.Errorf("req_duration samples tagged create_session: got %d, want 1", got)
	}
}

func TestClient_TryCreateSessionAppliesCplaneOverrideWithAutoTEID(t *testing.T) {
	pgw := startRefPGW(t)
	cli, _ := newRunClient(t)
	if _, err := cli.Connect(ConnectionOptions{Saddr: "127.0.0.1:0", Daddr: pgw, IfTypeName: "IFTypeS5S8SGWGTPC"}); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	override := uint8(gtpv2.IFTypeS11S4SGWGTPC)
	p := s5s8Params()
	p.IfTypeCplane = &override
	if r := cli.TryCreateSessionS5S8(pgw, p); !r.Ok || r.Cause != gtpv2.CauseRequestAccepted {
		t.Fatalf("TryCreateSessionS5S8: %+v", r)
	}
	sess, err := cli.Conn.GetSessionByIMSI(p.Imsi)
	if err != nil {
		t.Fatalf("GetSessionByIMSI: %v", err)
	}
	if _, err := sess.GetTEID(override); err != nil {
		t.Fatalf("session has no TEID under the overridden C-plane type: %v", err)
	}
	if r := cli.TryModifyBearerS5S8("", p); !r.Ok {
		t.Fatalf("TryModifyBearerS5S8 after override: %+v", r)
	}
}

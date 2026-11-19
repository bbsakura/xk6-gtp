package gtpv2

import (
	"testing"

	"github.com/wmnsk/go-gtp/gtpv2"
	"go.k6.io/k6/js/modulestest"
)

func TestS5S8SgwParams_IfTypeDefaults(t *testing.T) {
	var p S5S8SgwParams
	if got, want := p.cplaneIfType(), gtpv2.IFTypeS5S8SGWGTPC; got != want {
		t.Errorf("cplaneIfType default: got %d, want %d", got, want)
	}
	if got, want := p.uplaneIfType(), gtpv2.IFTypeS5S8SGWGTPU; got != want {
		t.Errorf("uplaneIfType default: got %d, want %d", got, want)
	}
	if got, want := p.cplanePeerIfType(), gtpv2.IFTypeS5S8PGWGTPC; got != want {
		t.Errorf("cplanePeerIfType default: got %d, want %d", got, want)
	}
}

func TestS5S8SgwParams_IfTypeOverrides(t *testing.T) {
	cplane, uplane, peer := gtpv2.IFTypeS11S4SGWGTPC, gtpv2.IFTypeS1USGWGTPU, gtpv2.IFTypeS11MMEGTPC
	p := S5S8SgwParams{
		IfTypeCplane:     &cplane,
		IfTypeUplane:     &uplane,
		IfTypeCplanePeer: &peer,
	}
	if got, want := p.cplaneIfType(), gtpv2.IFTypeS11S4SGWGTPC; got != want {
		t.Errorf("cplaneIfType override: got %d, want %d", got, want)
	}
	if got, want := p.uplaneIfType(), gtpv2.IFTypeS1USGWGTPU; got != want {
		t.Errorf("uplaneIfType override: got %d, want %d", got, want)
	}
	if got, want := p.cplanePeerIfType(), gtpv2.IFTypeS11MMEGTPC; got != want {
		t.Errorf("cplanePeerIfType override: got %d, want %d", got, want)
	}
}

func TestS5S8SgwParams_ExplicitZeroIfTypeIsAnOverride(t *testing.T) {
	zero := gtpv2.IFTypeS1UeNodeBGTPU
	p := S5S8SgwParams{IfTypeUplane: &zero}
	if got := p.uplaneIfType(); got != 0 {
		t.Errorf("uplaneIfType with explicit 0: got %d, want 0", got)
	}
}

func TestS5S8SgwParams_JSFieldNamesMapToOverrides(t *testing.T) {
	rt := modulestest.NewRuntime(t)
	var got S5S8SgwParams
	if err := rt.VU.RuntimeField.Set("capture", func(p S5S8SgwParams) { got = p }); err != nil {
		t.Fatal(err)
	}
	if _, err := rt.VU.RuntimeField.RunString(`capture({if_type_cplane: 10, if_type_uplane: 0})`); err != nil {
		t.Fatal(err)
	}
	if got.IfTypeCplane == nil || *got.IfTypeCplane != 10 {
		t.Errorf("if_type_cplane: got %v, want 10", got.IfTypeCplane)
	}
	if got.IfTypeUplane == nil || *got.IfTypeUplane != 0 {
		t.Errorf("if_type_uplane: got %v, want explicit 0", got.IfTypeUplane)
	}
	if got.IfTypeCplanePeer != nil {
		t.Errorf("if_type_cplane_peer: got %d, want unset", *got.IfTypeCplanePeer)
	}
}

func TestIfTypeExports_SanitySamples(t *testing.T) {
	m := ifTypeExports()
	for k, want := range map[string]uint8{
		"S5S8_SGW_GTPC":   gtpv2.IFTypeS5S8SGWGTPC,
		"S5S8_PGW_GTPC":   gtpv2.IFTypeS5S8PGWGTPC,
		"S11_MME_GTPC":    gtpv2.IFTypeS11MMEGTPC,
		"S11_S4_SGW_GTPC": gtpv2.IFTypeS11S4SGWGTPC,
	} {
		if got, ok := m[k]; !ok || got != want {
			t.Errorf("IFType[%q] = %d, %v; want %d, true", k, got, ok, want)
		}
	}
}

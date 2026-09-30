package gtpv2

import (
	"testing"

	"github.com/wmnsk/go-gtp/gtpv2"
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
	p := S5S8SgwParams{
		IfTypeCplane:     gtpv2.IFTypeS11S4SGWGTPC,
		IfTypeUplane:     gtpv2.IFTypeS1USGWGTPU,
		IfTypeCplanePeer: gtpv2.IFTypeS11MMEGTPC,
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

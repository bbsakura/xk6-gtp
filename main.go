package xk6gtp

import (
	"github.com/bbsakura/xk6-gtp/pkg/gtpv2"
	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/gtpv2", new(gtpv2.RootModule))
}

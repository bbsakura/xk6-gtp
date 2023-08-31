package xk6gtp

import (
	"go.k6.io/k6/js/modules"

	"github.com/bbsakura/xk6-gtp/pkg/gtpv2"
)

func init() {
	modules.Register("k6/x/gtpv2", new(gtpv2.RootModule))
}

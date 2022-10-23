package xk6gtp

import (
	"github.com/takehaya/xk6-gtp/pkg/gtpv2"
	"go.k6.io/k6/js/modules"
)

const version = "v0.0.1"

func init() {
	modules.Register("k6/x/gtpv2", &gtpv2.K6GTPv2{Version: version})
}

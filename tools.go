//go:build tools

package tools

import (
	_ "github.com/dmarkham/enumer"
	_ "github.com/elisescu/tty-share"
	_ "github.com/erkanzileli/co-author"
	_ "go.k6.io/xk6/cmd/xk6"
	_ "golang.org/x/tools/cmd/goimports"
	_ "golang.org/x/tools/cmd/stringer"
	_ "honnef.co/go/tools/cmd/staticcheck"
)

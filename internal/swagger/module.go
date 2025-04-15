// Package swagger
package swagger //nolint: predeclared

import (
	"github.com/nurdsoft/nurd-commerce-core/internal/swagger/transport/http"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
	"go.uber.org/fx"
)

// ModuleParams for swagger.
type ModuleParams struct {
	fx.In

	HTTPServer *httpTransport.Server
}

// NewModule for swagger.
// nolint:gocritic
func NewModule(p ModuleParams) error {

	http.RegisterTransport(p.HTTPServer)

	return nil
}

var (
	// ModuleServeSwagger for uber fx.
	ModuleServeSwagger = fx.Options(fx.Invoke(NewModule))
)

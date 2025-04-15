// Package transport
package transport

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nurdsoft/nurd-commerce-core/internal/transport/service"
)

type AccessControlAllowOrigins []string

// ModuleParams for transport.
type ModuleParams struct {
	fx.In

	Logger                    *zap.SugaredLogger
	AccessControlAllowOrigins AccessControlAllowOrigins
}

// NewModule
// nolint:gocritic
func NewModule(p ModuleParams) (Client, error) {
	return NewClient(service.New(p.AccessControlAllowOrigins), p.Logger), nil
}

var (
	// ModuleAPI for uber fx.
	ModuleAPI = fx.Options(fx.Provide(NewModule))
)

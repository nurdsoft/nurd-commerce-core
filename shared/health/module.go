// Package health allows the application to perform health checks
package health

import (
	"github.com/nurdsoft/nurd-commerce-core/shared/health/check"
	"github.com/nurdsoft/nurd-commerce-core/shared/health/endpoint"
	"github.com/nurdsoft/nurd-commerce-core/shared/health/service"
	"github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
	"go.uber.org/fx"

	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/health/transport/http"
)

// NewModule health checks
func NewModule(server *http.Server, checkers []check.Checker) {
	healthService := service.NewService(checkers)
	healthEndpoints := endpoint.New(healthService)

	httpTransport.RegisterTransport(server, healthEndpoints)
}

// Module for uber fx
var Module = fx.Options(
	fx.Invoke(
		NewModule,
	),
)

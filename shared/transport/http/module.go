// Package http contains http client/server with all necessary interceptor for logging, tracing, etc
package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ClientModuleParams contain client module params
type ClientModuleParams struct {
	fx.In

	Config cfg.Config
	Logger *zap.SugaredLogger
}

// NewClientModule returns new client module
// nolint:gocritic
func NewClientModule(p ClientModuleParams) *http.Client {
	return NewClient(
		WithLogger(p.Logger),
		WithUserAgent(p.Config.UserAgent),
	)
}

// ServerModuleParams contain server module params
type ServerModuleParams struct {
	fx.In

	Config Config
	Logger *zap.SugaredLogger
}

// NewServerModule returns new module for uber fx
// nolint:gocritic
func NewServerModule(lc fx.Lifecycle, s fx.Shutdowner, p ServerModuleParams) *Server {
	server := NewServerModuleWithoutLifecycle(p)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			p.Logger.Infow("starting HTTP server", "port", p.Config.Port)

			go func() {
				defer s.Shutdown() //nolint:errcheck

				if err := server.Serve(); err != nil {
					if err != http.ErrServerClosed {
						p.Logger.Errorw("serve HTTP server", "error", err)
					}
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("stopping HTTP server")

			return server.Stop(ctx)
		},
	})

	return server
}

// NewServerModuleWithoutLifecycle returns new module for uber fx without lifecycle hooks
// nolint:gocritic
func NewServerModuleWithoutLifecycle(p ServerModuleParams) *Server {
	opts := []Option{
		WithPrometheus(true),
		WithLogger(p.Logger),
	}

	return NewServer(fmt.Sprintf(":%d", p.Config.Port), opts...)
}

// Modules for uber fx
var (
	Module = fx.Options(
		fx.Provide(
			NewClientModule,
			NewServerModule,
		),
	)
	ModuleWithoutLifecycle = fx.Options(
		fx.Provide(
			NewClientModule,
			NewServerModuleWithoutLifecycle,
		),
	)
)

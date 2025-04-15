// Package log is based on uber zap
package log

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ModuleParams contain dependencies for module
type ModuleParams struct {
	fx.In

	ConfigCommon cfg.Config
	Config       Config
}

// NewModule returns module for uber fx
//
//nolint:gocritic
func NewModule(lc fx.Lifecycle, params ModuleParams) (*zap.SugaredLogger, error) {
	logger, err := New(
		WithFileLogEnabled(params.Config.FileLogEnabled))
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Infow("starting...", "env", params.ConfigCommon.Env, "version", params.ConfigCommon.Version)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// TODO: fix Error: sync /dev/stderr: inappropriate ioctl for device
			_ = logger.Sync() //nolint:errcheck
			return nil
		},
	})

	return logger, err
}

// Module for uber fx
var Module = fx.Options(
	fx.Provide(
		NewModule,
	),
)

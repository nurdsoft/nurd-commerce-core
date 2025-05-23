package ups

import (
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/ups/config"
	"net/http"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/ups/client"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/ups/service"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ModuleParams contain dependencies for module
type ModuleParams struct {
	fx.In

	Config     config.Config
	HttpClient *http.Client
	Logger     *zap.SugaredLogger
}

// NewModule
// nolint:gocritic
func NewModule(p ModuleParams) (client.Client, error) {
	svc, err := service.New(p.HttpClient, p.Config, p.Logger)
	if err != nil {
		return nil, err
	}

	client := client.NewClient(svc)

	return client, nil
}

var (
	// Module for uber fx.
	Module = fx.Options(fx.Provide(NewModule))
)

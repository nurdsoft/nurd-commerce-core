package webhook

import (
	"net/http"

	"github.com/nurdsoft/nurd-commerce-core/internal/webhook/client"
	"github.com/nurdsoft/nurd-commerce-core/internal/webhook/config"
	"github.com/nurdsoft/nurd-commerce-core/internal/webhook/service"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ModuleParams for user.
type ModuleParams struct {
	fx.In

	Config     config.Config
	Logger     *zap.SugaredLogger
	HttpClient *http.Client
}

// NewModule for redesign.
// nolint:gocritic
func NewModule(p ModuleParams) (client.Client, error) {
	svc := service.New(p.Logger, p.Config, p.HttpClient)
	client := client.NewClient(svc)

	return client, nil
}

var (
	// Module for uber fx.
	Module = fx.Options(fx.Provide(NewModule))
)

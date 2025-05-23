package shipping

import (
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/client"
	shipengineClient "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/shipengine/client"
	shipengineService "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/shipengine/service"
	upsClient "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/ups/client"
	upsService "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/ups/service"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
)

// ModuleParams contain dependencies for module
type ModuleParams struct {
	fx.In

	Config     Config
	HttpClient *http.Client
	Logger     *zap.SugaredLogger
}

// NewModule
// nolint:gocritic
func NewModule(p ModuleParams) (client.Client, error) {
	switch p.Config.Provider {
	case ProviderShipengine, "":
		service, err := shipengineService.New(p.HttpClient, p.Config.Shipengine, p.Logger)
		if err != nil {
			return nil, err
		}

		return shipengineClient.NewClient(service), nil
	case ProviderUPS:
		service, err := upsService.New(p.HttpClient, p.Config.UPS, p.Logger)
		if err != nil {
			return nil, err
		}

		return upsClient.NewClient(service), nil
	default:
		return nil, errors.Errorf("unknown provider: %s", p.Config.Provider)
	}
}

var (
	// Module for uber fx.
	Module = fx.Options(fx.Provide(NewModule))
)

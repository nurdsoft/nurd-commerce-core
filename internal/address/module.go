package address

import (
	"database/sql"

	"github.com/nurdsoft/nurd-commerce-core/internal/address/service"
	"github.com/nurdsoft/nurd-commerce-core/internal/address/transport/http"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/customerclient"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	salesforce "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	shipengine "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/shipengine/client"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nurdsoft/nurd-commerce-core/internal/address/endpoints"
	"github.com/nurdsoft/nurd-commerce-core/internal/address/repository"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
)

// ModuleParams for address.
type ModuleParams struct {
	fx.In

	DB               *sql.DB
	GormDB           *gorm.DB
	HTTPServer       *httpTransport.Server
	APPTransport     svcTransport.Client
	CommonConfig     cfg.Config
	ShipengineClient shipengine.Client
	Logger           *zap.SugaredLogger
	SalesforceClient salesforce.Client
	CustomerClient   customerclient.Client
}

// NewModule
// nolint:gocritic
func NewModule(p ModuleParams) error {
	repo := repository.New(p.DB, p.GormDB)
	svc := service.New(repo, p.Logger, p.CommonConfig, p.ShipengineClient, p.SalesforceClient, p.CustomerClient)
	eps := endpoints.New(svc)

	http.RegisterTransport(p.HTTPServer, eps, p.APPTransport)

	return nil
}

var (
	// ModuleHttpAPI for uber fx.
	ModuleHttpAPI = fx.Options(fx.Invoke(NewModule))
)

package cart

import (
	"database/sql"
	"github.com/nurdsoft/nurd-commerce-core/internal/address/addressclient"

	"github.com/nurdsoft/nurd-commerce-core/internal/cart/service"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/transport/http"
	"github.com/nurdsoft/nurd-commerce-core/shared/cache"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	shipengine "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/shipengine/client"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nurdsoft/nurd-commerce-core/internal/cart/endpoints"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/repository"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/productclient"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
	salesforce "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	stripe "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/client"
)

// ModuleParams for cart.
type ModuleParams struct {
	fx.In

	DB               *sql.DB
	GormDB           *gorm.DB
	HTTPServer       *httpTransport.Server
	APPTransport     svcTransport.Client
	CommonConfig     cfg.Config
	Logger           *zap.SugaredLogger
	ShipengineClient shipengine.Client
	StripeClient     stripe.Client
	ProductClient    productclient.Client
	AddressClient    addressclient.Client
	SalesforceClient salesforce.Client
}

// NewModule
// nolint:gocritic
func NewModule(p ModuleParams) error {
	repo := repository.New(p.DB, p.GormDB)
	cacheClient := cache.New()
	svc := service.New(repo, p.Logger, p.ShipengineClient, p.StripeClient, cacheClient, p.ProductClient, p.AddressClient, p.SalesforceClient)
	eps := endpoints.New(svc)

	http.RegisterTransport(p.HTTPServer, eps, p.APPTransport)

	return nil
}

var (
	// ModuleHttpAPI for uber fx.
	ModuleHttpAPI = fx.Options(fx.Invoke(NewModule))
)

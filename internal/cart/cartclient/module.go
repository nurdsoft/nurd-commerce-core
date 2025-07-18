package cartclient

import (
	"database/sql"

	"github.com/nurdsoft/nurd-commerce-core/internal/cart/service"
	"github.com/nurdsoft/nurd-commerce-core/shared/cache"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nurdsoft/nurd-commerce-core/internal/address/addressclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/repository"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/productclient"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory"
	salesforce "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	shipping "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/client"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes"
)

// ModuleParams for cartclient.
type ModuleParams struct {
	fx.In

	DB               *sql.DB
	GormDB           *gorm.DB
	HTTPServer       *httpTransport.Server
	APPTransport     svcTransport.Client
	CommonConfig     cfg.Config
	Logger           *zap.SugaredLogger
	ShippingClient   shipping.Client
	TaxesClient      taxes.Client
	ProductClient    productclient.Client
	AddressClient    addressclient.Client
	SalesforceClient salesforce.Client
	InventoryClient  inventory.Client
}

// NewClientModule
// nolint:gocritic
func NewClientModule(p ModuleParams) Client {
	repo := repository.New(p.DB, p.GormDB)
	cacheClient := cache.NewMemoryCache()
	svc := service.New(repo, p.Logger, p.ShippingClient, p.TaxesClient, cacheClient, p.ProductClient, p.AddressClient, p.InventoryClient, p.SalesforceClient)

	client := NewClient(svc)

	return client
}

var (
	// ModuleClient for uber fx.
	ModuleClient = fx.Options(fx.Provide(NewClientModule))
)

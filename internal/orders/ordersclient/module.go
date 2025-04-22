package ordersclient

import (
	"database/sql"
	"github.com/nurdsoft/nurd-commerce-core/internal/address/addressclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/customerclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/productclient"
	webhookClient "github.com/nurdsoft/nurd-commerce-core/internal/webhook/client"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/wishlistclient"

	"github.com/nurdsoft/nurd-commerce-core/internal/orders/service"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	salesforce "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	cart "github.com/nurdsoft/nurd-commerce-core/internal/cart/cartclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/repository"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
	stripe "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/client"
)

// ModuleParams for user.
type ModuleParams struct {
	fx.In

	DB               *sql.DB
	GormDB           *gorm.DB
	HTTPServer       *httpTransport.Server
	APPTransport     svcTransport.Client
	CommonConfig     cfg.Config
	Logger           *zap.SugaredLogger
	CartClient       cart.Client
	StripeClient     stripe.Client
	WishlistClient   wishlistclient.Client
	SalesforceClient salesforce.Client
	CustomerClient   customerclient.Client
	AddressClient    addressclient.Client
	ProductClient    productclient.Client
	WebhookClient    webhookClient.Client
}

// NewClientModule
// nolint:gocritic
func NewClientModule(p ModuleParams) Client {
	repo := repository.New(p.DB, p.GormDB)
	svc := service.New(
		repo, p.Logger, p.CustomerClient, p.CartClient, p.StripeClient,
		p.WishlistClient, p.CommonConfig, p.SalesforceClient, p.AddressClient, p.ProductClient, p.WebhookClient)

	client := NewClient(svc)

	return client
}

var (
	// ModuleClient for uber fx.
	ModuleClient = fx.Options(fx.Provide(NewClientModule))
)

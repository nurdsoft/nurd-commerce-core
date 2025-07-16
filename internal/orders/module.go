package orders

import (
	"database/sql"

	"github.com/nurdsoft/nurd-commerce-core/internal/address/addressclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/service"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/transport/http"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/productclient"
	webhookClient "github.com/nurdsoft/nurd-commerce-core/internal/webhook/client"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/wishlistclient"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	cart "github.com/nurdsoft/nurd-commerce-core/internal/cart/cartclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/customerclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/endpoints"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/repository"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
)

// ModuleParams for user.
type ModuleParams struct {
	fx.In

	DB              *sql.DB
	GormDB          *gorm.DB
	HTTPServer      *httpTransport.Server
	APPTransport    svcTransport.Client
	CommonConfig    cfg.Config
	Logger          *zap.SugaredLogger
	CustomerClient  customerclient.Client
	CartClient      cart.Client
	PaymentClient   payment.Client
	WishlistClient  wishlistclient.Client
	InventoryClient inventory.Client
	AddressClient   addressclient.Client
	ProductClient   productclient.Client
	WebhookClient   webhookClient.Client
}

// NewModule for redesign.
// nolint:gocritic
func NewModule(p ModuleParams) error {
	repo := repository.New(p.DB, p.GormDB)
	svc := service.New(repo, p.Logger, p.CustomerClient, p.CartClient, p.PaymentClient, p.WishlistClient,
		p.CommonConfig, p.InventoryClient, p.AddressClient, p.ProductClient, p.WebhookClient)
	eps := endpoints.New(svc)

	http.RegisterTransport(p.HTTPServer, eps, p.APPTransport)

	return nil
}

var (
	// ModuleHttpAPI for uber fx.
	ModuleHttpAPI = fx.Options(fx.Invoke(NewModule))
)

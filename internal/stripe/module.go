package stripe

import (
	"database/sql"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/customerclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/ordersclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/stripe/service"
	"github.com/nurdsoft/nurd-commerce-core/internal/stripe/transport/http"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	stripeClient "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/client"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nurdsoft/nurd-commerce-core/internal/stripe/endpoints"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
)

// ModuleParams for customer.
type ModuleParams struct {
	fx.In

	DB             *sql.DB
	GormDB         *gorm.DB
	HTTPServer     *httpTransport.Server
	APPTransport   svcTransport.Client
	CommonConfig   cfg.Config
	Logger         *zap.SugaredLogger
	OrdersClient   ordersclient.Client
	StripeClient   stripeClient.Client
	CustomerClient customerclient.Client
}

// NewModule
// nolint:gocritic
func NewModule(p ModuleParams) error {
	svc := service.New(p.Logger, p.CommonConfig, p.StripeClient, p.OrdersClient, p.CustomerClient)
	eps := endpoints.New(svc)

	http.RegisterTransport(p.HTTPServer, eps, p.APPTransport)

	return nil
}

var (
	// ModuleHttpAPI for uber fx.
	ModuleHttpAPI = fx.Options(fx.Invoke(NewModule))
)

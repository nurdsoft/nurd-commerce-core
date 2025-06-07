package authorizenet

import (
	"database/sql"

	"github.com/nurdsoft/nurd-commerce-core/internal/authorizenet/service"
	"github.com/nurdsoft/nurd-commerce-core/internal/authorizenet/transport/http"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/customerclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/ordersclient"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment"
	authorizenetClient "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/client"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nurdsoft/nurd-commerce-core/internal/authorizenet/endpoints"
	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
)

// ModuleParams for customer.
type ModuleParams struct {
	fx.In

	DB                 *sql.DB
	GormDB             *gorm.DB
	HTTPServer         *httpTransport.Server
	APPTransport       svcTransport.Client
	CommonConfig       cfg.Config
	PaymentConfig      payment.Config
	Logger             *zap.SugaredLogger
	OrdersClient       ordersclient.Client
	AuthorizeNetClient authorizenetClient.Client
	CustomerClient     customerclient.Client
}

// NewModule
// nolint:gocritic
func NewModule(p ModuleParams) error {
	svc := service.New(p.Logger, p.CommonConfig, p.AuthorizeNetClient, p.OrdersClient, p.CustomerClient)
	eps := endpoints.New(svc)

	signatureKey := p.PaymentConfig.AuthorizeNet.SignatureKey

	http.RegisterTransport(p.HTTPServer, eps, p.APPTransport, signatureKey)

	return nil
}

var (
	// ModuleHttpAPI for uber fx.
	ModuleHttpAPI = fx.Options(fx.Invoke(NewModule))
)

// Package cmd contains commands
package cmd

import (
	"database/sql"
	"time"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe"

	"github.com/nurdsoft/nurd-commerce-core/config"
	"github.com/nurdsoft/nurd-commerce-core/internal/address"
	"github.com/nurdsoft/nurd-commerce-core/internal/address/addressclient"
	authorizenetModule "github.com/nurdsoft/nurd-commerce-core/internal/authorizenet"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/cartclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/customerclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/ordersclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/product"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/productclient"
	stripeModule "github.com/nurdsoft/nurd-commerce-core/internal/stripe"
	"github.com/nurdsoft/nurd-commerce-core/internal/swagger"
	"github.com/nurdsoft/nurd-commerce-core/internal/transport"
	"github.com/nurdsoft/nurd-commerce-core/internal/webhook"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/wishlistclient"
	"github.com/nurdsoft/nurd-commerce-core/shared/db"
	"github.com/nurdsoft/nurd-commerce-core/shared/health"
	"github.com/nurdsoft/nurd-commerce-core/shared/health/check"
	"github.com/nurdsoft/nurd-commerce-core/shared/log"
	"github.com/nurdsoft/nurd-commerce-core/shared/module"
	httpTransport "github.com/nurdsoft/nurd-commerce-core/shared/transport/http"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var apiCommand = &cobra.Command{
	Use:   "api",
	Short: "Serve API",
	RunE: func(cmd *cobra.Command, args []string) error {
		return module.Run(
			fx.Provide(
				config.New(cfgFile, version),
				func(db *sql.DB) []check.Checker { return []check.Checker{check.NewSQLChecker(db)} },
			),
			db.Module,
			httpTransport.Module,
			transport.ModuleAPI,
			health.Module,
			shipping.Module,
			stripe.Module,
			authorizenet.Module,
			payment.Module,
			taxes.Module,
			log.Module,
			customer.ModuleHttpAPI,
			customerclient.ModuleClient,
			product.ModuleHttpAPI,
			productclient.ModuleClient,
			wishlist.ModuleHttpAPI,
			wishlistclient.ModuleClient,
			address.ModuleHttpAPI,
			addressclient.ModuleClient,
			cart.ModuleHttpAPI,
			cartclient.ModuleClient,
			orders.ModuleHttpAPI,
			ordersclient.ModuleClient,
			webhook.Module,
			inventory.Module,
			salesforce.Module,
			swagger.ModuleServeSwagger,
			stripeModule.ModuleHttpAPI,
			authorizenetModule.ModuleHttpAPI,
			fx.NopLogger,
			fx.StartTimeout(time.Second*60),
		)
	},
}

func init() {
	rootCmd.AddCommand(apiCommand)
}

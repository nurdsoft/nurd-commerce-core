// Package payment contains function which helps to work with Payment
package payment

import (
	authorizenetConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/config"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/providers"
	stripeConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/config"
	"github.com/pkg/errors"
)

// Config should be included as part of service config.
type Config struct {
	Provider     providers.ProviderType
	Stripe       stripeConfig.Config
	AuthorizeNet authorizenetConfig.Config
}

// Validate config.
func (c *Config) Validate() error {
	switch c.Provider {
	case providers.ProviderStripe, "": // defaults to Stripe
		return c.Stripe.Validate()
	case providers.ProviderAuthorizeNet:
		return c.AuthorizeNet.Validate()
	default:
		return errors.Errorf("unknown provider: %s", c.Provider)
	}
}

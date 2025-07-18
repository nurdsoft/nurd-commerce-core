// Package taxes contains function which helps to work with Taxes
package taxes

import (
	"errors"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/providers"
	stripeConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/config"
)

// Config should be included as part of service config.
type Config struct {
	Provider providers.ProviderType
	Stripe   stripeConfig.Config
}

// Validate config.
func (c *Config) Validate() error {
	switch c.Provider {
	case "":
		return nil
	case providers.ProviderStripe:
		return c.Stripe.Validate()
	default:
		return errors.New("invalid tax provider")
	}
}

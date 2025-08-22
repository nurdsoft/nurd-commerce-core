// Package taxes contains function which helps to work with Taxes
package taxes

import (
	"errors"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/providers"
	stripeConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/config"
	taxjarConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/taxjar/config"
)

// Config should be included as part of service config.
type Config struct {
	Provider providers.ProviderType
	Stripe   stripeConfig.Config
	TaxJar   taxjarConfig.Config
}

// Validate config.
func (c *Config) Validate() error {
	switch c.Provider {
	case "", providers.ProviderNone:
		return nil
	case providers.ProviderStripe:
		return c.Stripe.Validate()
	case providers.ProviderTaxJar:
		return c.TaxJar.Validate()
	default:
		return errors.New("invalid tax provider")
	}
}

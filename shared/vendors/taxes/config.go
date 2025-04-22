// Package taxes contains function which helps to work with Taxes
package taxes

import (
	stripeConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/config"
)

// Config should be included as part of service config.
type Config struct {
	Stripe stripeConfig.Config
}

// Validate config.
func (c *Config) Validate() error {
	// TODO Add conditional to use validate() to enabled tax vendor
	return c.Stripe.Validate()
}

// Package payment contains function which helps to work with Payment
package payment

import (
	stripeConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/config"
)

// Config should be included as part of service config.
type Config struct {
	Stripe stripeConfig.Config
}

// Validate config.
func (c *Config) Validate() error {
	// TODO Add conditional to use validate() to enabled payment vendor
	return c.Stripe.Validate()
}

// Package inventory contains function which helps to work with Inventory
package inventory

import (
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/providers"
	salesforceConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/config"
)

// Config should be included as part of service config.
type Config struct {
	Salesforce salesforceConfig.Config
	Provider   providers.ProviderType
}

// Validate config.
func (c *Config) Validate() error {
	switch c.Provider {
	case providers.ProviderSalesforce:
		return c.Salesforce.Validate()
	default:
		return nil
	}
}

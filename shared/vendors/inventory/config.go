// Package inventory contains function which helps to work with Inventory
package inventory

import (
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/providers"
	printfulConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/printful/config"
	salesforceConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/config"
)

// Config should be included as part of service config.
type Config struct {
	Salesforce salesforceConfig.Config
	Printful   printfulConfig.Config
	Provider   providers.ProviderType
}

// Validate config.
func (c *Config) Validate() error {
	switch c.Provider {
	case providers.ProviderNone:
		return nil
	case providers.ProviderSalesforce:
		return c.Salesforce.Validate()
	case providers.ProviderPrintful:
		return c.Printful.Validate()
	default:
		return nil
	}
}

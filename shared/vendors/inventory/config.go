// Package inventory contains function which helps to work with Inventory
package inventory

import (
	salesforceConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/config"
)

// Config should be included as part of service config.
type Config struct {
	Salesforce salesforceConfig.Config
}

// Validate config.
func (c *Config) Validate() error {
	// TODO Add conditional to use validate() to enabled payment vendor
	return c.Salesforce.Validate()
}

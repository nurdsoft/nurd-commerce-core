// Package shipping contains function which helps to work with Shipping
package shipping

import (
	shipengineConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/shipengine/config"
)

// Config should be included as part of service config.
type Config struct {
	Shipengine shipengineConfig.Config
}

// Validate config.
func (c *Config) Validate() error {
	// TODO Add conditional to use validate() to enabled shipping vendor
	return c.Shipengine.Validate()
}

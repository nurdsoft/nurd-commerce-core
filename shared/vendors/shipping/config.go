// Package shipping contains function which helps to work with Shipping
package shipping

import (
	shipengineConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/shipengine/config"
	upsConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/ups/config"
	"github.com/pkg/errors"
)

const (
	ProviderNone       string = "none"
	ProviderShipengine string = "Shipengine"
	ProviderUPS        string = "UPS"
)

// Config should be included as part of service config.
type Config struct {
	Provider   string
	Shipengine shipengineConfig.Config
	UPS        upsConfig.Config
}

// Validate config.
func (c *Config) Validate() error {
	switch c.Provider {
	case "", ProviderNone:
		return nil
	case ProviderShipengine:
		return c.Shipengine.Validate()
	case ProviderUPS:
		return c.UPS.Validate()
	default:
		return errors.Errorf("unknown provider: %s", c.Provider)
	}
}

package config

import (
	"strings"

	svcTransport "github.com/nurdsoft/nurd-commerce-core/internal/transport"
	salesforce "github.com/nurdsoft/nurd-commerce-core/internal/vendors/salesforce/config"
	shipengine "github.com/nurdsoft/nurd-commerce-core/internal/vendors/shipengine/config"
	stripe "github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/config"
	webhook "github.com/nurdsoft/nurd-commerce-core/internal/webhook/config"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	"github.com/nurdsoft/nurd-commerce-core/shared/db"
	"github.com/nurdsoft/nurd-commerce-core/shared/log"
	"github.com/nurdsoft/nurd-commerce-core/shared/transport"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

// Config is represented main configuration of service
type Config struct {
	fx.Out

	Common                    cfg.Config
	Transport                 transport.Config
	Logger                    log.Config
	DB                        db.Config
	AccessControlAllowOrigins svcTransport.AccessControlAllowOrigins
	Shipengine                shipengine.Config
	Stripe                    stripe.Config
	Salesforce                salesforce.Config
	Webhook                   webhook.Config
}

// Validate config
func (c *Config) Validate() error {
	var errs []string

	validatables := []cfg.Validatable{
		&c.DB,
		&c.Common,
		&c.Shipengine,
		&c.Stripe,
		&c.Salesforce,
		&c.Webhook,
	}

	if err := cfg.ValidateConfigs(validatables...); err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, ","))
	}

	return nil
}

// New config.
func New(path, version string) func() (Config, error) {
	return func() (Config, error) {
		cfgFile := Config{}
		cfgFile.Common.Version = version

		err := cfg.Init("config", path, &cfgFile)
		if err != nil {
			return cfgFile, err
		}

		return cfgFile, nil
	}
}

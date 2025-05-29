package config

import (
	"strings"

	"github.com/pkg/errors"
)

// Config is a common service config
type Config struct {
	SecurityHost  string
	APIHost       string
	ClientID      string
	ClientSecret  string
	ShipperName string
	ShipperNumber string
}

// Validate config
func (c *Config) Validate() error {
	var errs []string

	if c.SecurityHost == "" {
		errs = append(errs, "ups securityhost shouldn't be empty")
	}

	if c.APIHost == "" {
		errs = append(errs, "ups apihost shouldn't be empty")
	}

	if c.ClientID == "" {
		errs = append(errs, "ups clientid shouldn't be empty")
	}

	if c.ClientSecret == "" {
		errs = append(errs, "ups clientsecret shouldn't be empty")
	}

	if c.ShipperName == "" {
		errs = append(errs, "ups shippername shouldn't be empty")
	}

	if c.ShipperNumber == "" {
		errs = append(errs, "ups shippernumber shouldn't be empty")
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, ","))
	}

	return nil
}

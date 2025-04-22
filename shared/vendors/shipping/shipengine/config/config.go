package config

import (
	"strings"

	"github.com/pkg/errors"
)

// Config is a common service config
type Config struct {
	Host       string
	Token      string
	CarrierIds string
}

// Validate config
func (c *Config) Validate() error {
	var errs []string

	if c.Host == "" {
		errs = append(errs, "shipengine apihost shouldn't be empty")
	}

	if c.Token == "" {
		errs = append(errs, "shipengine token shouldn't be empty")
	}

	if c.CarrierIds == "" {
		errs = append(errs, "shipengine carrierIds shouldn't be empty")
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, ","))
	}

	return nil
}

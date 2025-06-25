package config

import (
	"strings"

	"github.com/pkg/errors"
)

// Config is a common service config
type Config struct {
	ApiLoginID     string
	TransactionKey string
	LiveMode       bool
	Endpoint       string
	SignatureKey   string
}

// Validate config
func (c *Config) Validate() error {
	var errs []string

	if c.ApiLoginID == "" {
		errs = append(errs, "authorize.net api login ID shouldn't be empty")
	}

	if c.TransactionKey == "" {
		errs = append(errs, "authorize.net transaction key shouldn't be empty")
	}

	if c.SignatureKey == "" {
		errs = append(errs, "authorize.net signature key shouldn't be empty")
	}

	if c.Endpoint == "" {
		errs = append(errs, "authorize.net endpoint shouldn't be empty")
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, ","))
	}

	return nil
}

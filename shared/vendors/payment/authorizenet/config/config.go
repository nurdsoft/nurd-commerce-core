package config

import (
	"strings"

	"github.com/pkg/errors"
)

// Config is a common service config
type Config struct {
	ApiLoginID     string
	TransactionKey string
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

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, ","))
	}

	return nil
}

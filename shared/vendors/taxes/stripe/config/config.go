package config

import (
	"strings"

	"github.com/pkg/errors"
)

// Config is a common service config
type Config struct {
	Key string
}

// Validate config
func (c *Config) Validate() error {
	var errs []string

	if c.Key == "" {
		errs = append(errs, "stripe key shouldn't be empty")
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, ","))
	}

	return nil
}

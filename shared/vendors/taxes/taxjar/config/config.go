package config

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/taxjar/taxjar-go"
)

type Config struct {
	Key string
	URL string
}

// Validate config
func (c *Config) Validate() error {
	var errs []string

	if c.Key == "" {
		errs = append(errs, "taxjar api key shouldn't be empty")
	}

	if c.URL == "" {
		c.URL = taxjar.SandboxAPIURL
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, ","))
	}

	return nil
}

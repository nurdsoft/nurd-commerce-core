package config

import (
	"strings"

	"github.com/pkg/errors"
)

// Config is a common service config
type Config struct {
	OAuthToken string
	BaseURL    string
}

// Validate config
func (c *Config) Validate() error {
	var errs []string

	if c.OAuthToken == "" {
		errs = append(errs, "Printful OAuth token shouldn't be empty")
	}

	if c.BaseURL == "" {
		errs = append(errs, "Printful base URL shouldn't be empty")
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, ","))
	}

	return nil
}

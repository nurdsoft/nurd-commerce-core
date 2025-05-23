package config

import (
	"strings"

	"github.com/pkg/errors"
)

// Config is a common service config
type Config struct {
	SecurityHost string
	APIHost      string
	ClientID     string
	ClientSecret string
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

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, ","))
	}

	return nil
}

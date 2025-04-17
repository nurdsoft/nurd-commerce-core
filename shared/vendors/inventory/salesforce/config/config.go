package config

import (
	"strings"

	"github.com/pkg/errors"
)

// Config is a common service config
type Config struct {
	ApiHost      string
	ApiVersion   string
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
}

// Validate config
func (c *Config) Validate() error {
	var errs []string

	if c.ApiHost == "" {
		errs = append(errs, "Salesforce API host shouldn't be empty")
	}

	if c.ApiVersion == "" {
		c.ApiVersion = "v62.0"
	}

	if c.ClientID == "" {
		errs = append(errs, "Salesforce client ID shouldn't be empty")
	}

	if c.ClientSecret == "" {
		errs = append(errs, "Salesforce client secret shouldn't be empty")
	}

	if c.Username == "" {
		errs = append(errs, "Salesforce username shouldn't be empty")
	}

	if c.Password == "" {
		errs = append(errs, "Salesforce password shouldn't be empty")
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, ","))
	}

	return nil
}

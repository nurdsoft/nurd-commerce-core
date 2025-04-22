package config

import (
	"strings"

	"github.com/pkg/errors"
)

// Config is a common service config
type Config struct {
	OrderURL string
	Token    string
}

// Validate config
func (c *Config) Validate() error {
	var errs []string

	if c.OrderURL == "" {
		errs = append(errs, "webhook orderURL host shouldn't be empty")
	}

	if c.Token == "" {
		errs = append(errs, "webhook token shouldn't be empty")
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, ","))
	}

	return nil
}

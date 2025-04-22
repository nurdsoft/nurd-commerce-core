package cfg

import (
	"strings"

	"github.com/pkg/errors"
)

// Config is a common service config
type Config struct {
	Name      string
	Env       string
	Version   string
	UserAgent string
}

// Validate config
func (c *Config) Validate() error {
	var errs []string

	if c.Name == "" {
		errs = append(errs, "Name shouldn't be empty")
	}

	if c.Env == "" {
		errs = append(errs, "Env shouldn't be empty")
	}

	if c.Version == "" {
		errs = append(errs, "Version shouldn't be empty")
	}

	if c.UserAgent == "" {
		errs = append(errs, "UserAgent shouldn't be empty")
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, ","))
	}

	return nil
}

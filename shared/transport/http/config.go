// Package http contains http client/server with all necessary interceptor for logging, tracing, etc
package http

import (
	"strings"

	"github.com/pkg/errors"
)

// Config defines how we run server
type Config struct {
	Port int
}

// Validate config
func (c *Config) Validate() error {
	var errs []string

	if c.Port <= 0 {
		errs = append(errs, "Port shouldn't be empty")
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, ","))
	}

	return nil
}

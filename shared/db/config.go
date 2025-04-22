// Package db contains function which helps to work with PostgreSQL
package db

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// Postgres defines configuration of the database
type Postgres struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

// URL for postgres.
func (p *Postgres) URL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User,
		url.QueryEscape(p.Password),
		p.Host,
		p.Port,
		p.Database,
		p.SSLMode,
	)
}

// Validate config.
func (p *Postgres) Validate() error {
	var errs []string

	if p.Database == "" {
		errs = append(errs, "Database shouldn't be empty")
	}

	if p.Host == "" {
		errs = append(errs, "Host shouldn't be empty")
	}

	if p.Port <= 0 {
		errs = append(errs, "Port shouldn't be empty")
	}

	if p.User == "" {
		errs = append(errs, "User shouldn't be empty")
	}

	if p.Password == "" {
		errs = append(errs, "Password shouldn't be empty")
	}

	if p.SSLMode == "" {
		errs = append(errs, "SSLMode shouldn't be empty")
	}

	if len(errs) > 0 {
		return errors.Errorf("%s", strings.Join(errs, ","))
	}

	return nil
}

// Config should be included as part of service config.
type Config struct {
	Postgres Postgres
}

// Validate config.
func (c *Config) Validate() error {
	return c.Postgres.Validate()
}

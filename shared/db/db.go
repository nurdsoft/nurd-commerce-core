// Package db contains function which helps to work with PostgreSQL
package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib" // import postgres driver
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const maxOpenConnects = 20

// options for DB
type options struct {
	maxOpenConns int
}

// WithMaxOpenConns set max open connections to database
func WithMaxOpenConns(n int) Option {
	return optionFunc(func(o *options) {
		o.maxOpenConns = n
	})
}

// Option applies option
type Option interface{ apply(*options) }
type optionFunc func(*options)

func (f optionFunc) apply(o *options) { f(o) }

// New inits new database
func New(cfg *Config, opts ...Option) (*sql.DB, *gorm.DB, error) {
	// Default o
	o := options{
		maxOpenConns: maxOpenConnects,
	}

	for _, op := range opts {
		op.apply(&o)
	}

	driverName := "pgx"

	db, err := sql.Open(driverName, cfg.Postgres.URL())
	if err != nil {
		return nil, nil, errors.Wrap(err, "open connection")
	}

	if err := db.Ping(); err != nil {
		return nil, nil, errors.Wrap(err, "ping")
	}

	db.SetMaxOpenConns(o.maxOpenConns)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		return nil, nil, errors.Wrap(err, "gorm open connection")
	}

	return db, gormDB, nil
}

// NewGormDB inits new gorm database connection
func NewGormDB(cfg *Config, opts ...Option) (*sql.DB, error) {
	// Default o
	o := options{
		maxOpenConns: maxOpenConnects,
	}

	for _, op := range opts {
		op.apply(&o)
	}

	driverName := "pgx"

	db, err := sql.Open(driverName, cfg.Postgres.URL())
	if err != nil {
		return nil, errors.Wrap(err, "open connection")
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "ping")
	}

	db.SetMaxOpenConns(o.maxOpenConns)

	return db, nil
}

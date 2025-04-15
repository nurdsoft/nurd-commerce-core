// Package db contains function which helps to work with PostgreSQL
package db

import (
	"database/sql"

	"gorm.io/gorm"

	"go.uber.org/fx"
)

// ModuleParams contains module params
type ModuleParams struct {
	fx.In

	Config Config
}

// NewModule returns new module for uber fx
//
//nolint:gocritic
func NewModule(p ModuleParams) (*sql.DB, *gorm.DB, error) {
	return New(
		&p.Config,
	)
}

// Module for uber fx
var Module = fx.Options(
	fx.Provide(
		NewModule,
	),
)

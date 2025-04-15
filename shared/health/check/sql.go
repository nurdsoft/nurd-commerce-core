// Package check provides a way to do health checks
package check

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

var (
	// ErrCouldNotPing for SQL.
	ErrCouldNotPing = errors.New("could not ping")

	// ErrCouldNotQuery for SQL.
	ErrCouldNotQuery = errors.New("could not query")

	// ErrCouldNotGetRows for SQL.
	ErrCouldNotGetRows = errors.New("could not get rows")

	// ErrCouldNotScan for SQL.
	ErrCouldNotScan = errors.New("could not scan")
)

type sqlChecker struct {
	db *sql.DB
}

// NewSQLChecker for health checks
func NewSQLChecker(db *sql.DB) Checker {
	return &sqlChecker{db}
}

// Check the health of SQL.
func (c sqlChecker) Check(ctx context.Context) error {
	if err := c.db.PingContext(ctx); err != nil {
		return errors.Wrap(ErrCouldNotPing, err.Error())
	}

	rows, err := c.db.QueryContext(ctx, "SELECT 1")
	if err != nil {
		return errors.Wrap(ErrCouldNotQuery, err.Error())
	}

	err = rows.Err()
	if err != nil {
		return errors.Wrap(ErrCouldNotQuery, err.Error())
	}

	defer rows.Close()

	if !rows.Next() {
		return errors.Wrap(ErrCouldNotGetRows, "no rows")
	}

	var ok string
	if err := rows.Scan(&ok); err != nil {
		return errors.Wrap(ErrCouldNotScan, err.Error())
	}

	return nil
}

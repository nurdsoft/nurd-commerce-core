// Package db contains function which helps to work with PostgreSQL
package db

import (
	"database/sql"
	"strings"

	"gorm.io/gorm"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/pkg/errors"
)

// IsContextCanceledError shows if operation was canceled
func IsContextCanceledError(err error) bool {
	return strings.Contains(err.Error(), "canceled") // https://github.com/golang/go/issues/36208
}

// IsAlreadyExistError shows if entity already exist in db
func IsAlreadyExistError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
}

// IsForeignKeyViolationError shows if foreign key violation error
func IsForeignKeyViolationError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
		return true
	}

	if errors.Is(err, gorm.ErrForeignKeyViolated) {
		return true
	}

	return false
}

// IsInvalidValueError shows if foreign key violation error
func IsInvalidValueError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.InvalidTextRepresentation
}

// IsNotFoundError shows if foreign key violation error
func IsNotFoundError(err error) bool {
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}

	// Optional fallback in case some libraries return plain strings
	return err != nil && err.Error() == "record not found"
}

// IsInvalidValueError check if any field has invalid length
func IsInvalidLength(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.StringDataRightTruncationDataException {
		return true
	}
	return false
}

func IsUniqueViolationError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return true
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}

	return false
}

// Useful for invalid enum values
func IsInvalidTextRepresentationError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.InvalidTextRepresentation {
		return true
	}
	return false
}

func HandleDbError(err error) error {
	switch {
	case IsNotFoundError(err):
		return appErrors.NewAPIError("RECORD_NOT_FOUND")
	case IsUniqueViolationError(err):
		return appErrors.NewAPIError("DUPLICATED_KEY")
	case IsForeignKeyViolationError(err):
		return appErrors.NewAPIError("FOREIGN_KEY_VIOLATION")
	default:
		return err
	}
}

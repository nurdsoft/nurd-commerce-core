// Package db contains function which helps to work with PostgreSQL
package db

import (
	"database/sql"
	"strings"

	"gorm.io/gorm"

	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/pkg/errors"
)

// IsContextCanceledError shows if operation was canceled
func IsContextCanceledError(err error) bool {
	return strings.Contains(err.Error(), "canceled") // https://github.com/golang/go/issues/36208
}

// IsAlreadyExistError shows if entity already exist in db
func IsAlreadyExistError(err error) bool {
	if err, ok := err.(*pgconn.PgError); ok && err.Code == pgerrcode.UniqueViolation {
		return true
	}

	errCause := errors.Cause(err)
	if errCause, ok := errCause.(*pgconn.PgError); ok && errCause.Code == pgerrcode.UniqueViolation {
		return true
	}

	return false
}

// IsForeignKeyViolationError shows if foreign key violation error
func IsForeignKeyViolationError(err error) bool {
	if err, ok := err.(*pgconn.PgError); ok && err.Code == pgerrcode.ForeignKeyViolation {
		return true
	}

	errCause := errors.Cause(err)
	if errCause, ok := errCause.(*pgconn.PgError); ok && errCause.Code == pgerrcode.ForeignKeyViolation {
		return true
	}

	if errors.Is(err, gorm.ErrForeignKeyViolated) {
		return true
	}

	return false
}

// IsInvalidValueError shows if foreign key violation error
func IsInvalidValueError(err error) bool {
	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.InvalidTextRepresentation {
		return true
	}

	errCause := errors.Cause(err)
	if errC, ok := errCause.(*pgconn.PgError); ok && errC.Code == pgerrcode.InvalidTextRepresentation {
		return true
	}

	return false
}

// IsNotFoundError shows if foreign key violation error
func IsNotFoundError(err error) bool {
	if err == sql.ErrNoRows || err.Error() == "record not found" {
		return true
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}

	errCause := errors.Cause(err)

	return errCause == sql.ErrNoRows
}

// IsInvalidValueError check if any field has invalid length
func IsInvalidLength(err error) bool {
	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.StringDataRightTruncationDataException {
		return true
	}

	errCause := errors.Cause(err)
	if errC, ok := errCause.(*pgconn.PgError); ok && errC.Code == pgerrcode.StringDataRightTruncationDataException {
		return true
	}

	return false
}

func IsUniqueViolationError(err error) bool {
	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.UniqueViolation {
		return true
	}

	errCause := errors.Cause(err)
	if errC, ok := errCause.(*pgconn.PgError); ok && errC.Code == pgerrcode.UniqueViolation {
		return true
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	return false
}

// Useful for invalid enum values
func IsInvalidTextRepresentationError(err error) bool {
	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.InvalidTextRepresentation {
		return true
	}

	errCause := errors.Cause(err)
	if errC, ok := errCause.(*pgconn.PgError); ok && errC.Code == pgerrcode.InvalidTextRepresentation {
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

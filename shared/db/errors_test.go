package db

import (
	"database/sql"
	"errors"
	"testing"

	pkgErrors "github.com/pkg/errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	// pgError "github.com/jackc/pgconn"
)

func TestIsContextCanceledError(t *testing.T) {
	err := errors.New("context canceled")
	assert.True(t, IsContextCanceledError(err))

	err = errors.New("some other error")
	assert.False(t, IsContextCanceledError(err))
}

func TestIsAlreadyExistError(t *testing.T) {
	err := &pgconn.PgError{Code: pgerrcode.UniqueViolation}
	assert.True(t, IsAlreadyExistError(err))
}

func TestIsForeignKeyViolationError(t *testing.T) {
	err := &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation}
	assert.True(t, IsForeignKeyViolationError(err))
}

func TestIsInvalidValueError(t *testing.T) {
	err := &pgconn.PgError{Code: pgerrcode.InvalidTextRepresentation}
	assert.True(t, IsInvalidValueError(err))
}

func TestIsNotFoundError(t *testing.T) {
	err := sql.ErrNoRows
	assert.True(t, IsNotFoundError(err))

	err = gorm.ErrRecordNotFound
	assert.True(t, IsNotFoundError(err))
}

func TestIsInvalidLength(t *testing.T) {
	err := &pgconn.PgError{Code: pgerrcode.StringDataRightTruncationDataException}
	assert.True(t, IsInvalidLength(err))

	wrappedErr := pkgErrors.Wrap(err, "wrapped error")
	assert.True(t, IsInvalidLength(wrappedErr))
}

func TestIsUniqueViolationError(t *testing.T) {
	err := &pgconn.PgError{Code: pgerrcode.UniqueViolation}
	assert.True(t, IsUniqueViolationError(err))
}

func TestIsInvalidTextRepresentationError(t *testing.T) {
	err := &pgconn.PgError{Code: pgerrcode.InvalidTextRepresentation}
	assert.True(t, IsInvalidTextRepresentationError(err))

	wrappedErr := pkgErrors.Wrap(err, "wrapped error")
	assert.True(t, IsInvalidTextRepresentationError(wrappedErr))

	err = &pgconn.PgError{Message: "some other error"}
	assert.False(t, IsInvalidTextRepresentationError(err))
}

func TestHandleDbError(t *testing.T) {
	err := sql.ErrNoRows
	assert.Equal(t, "RECORD_NOT_FOUND", HandleDbError(err).(*appErrors.APIError).ErrorCode)

	err = &pgconn.PgError{Code: pgerrcode.UniqueViolation}
	assert.Equal(t, "DUPLICATED_KEY", HandleDbError(err).(*appErrors.APIError).ErrorCode)

	err = &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation}
	assert.Equal(t, "FOREIGN_KEY_VIOLATION", HandleDbError(err).(*appErrors.APIError).ErrorCode)

	err = errors.New("some other error")
	assert.Equal(t, err, HandleDbError(err))
}

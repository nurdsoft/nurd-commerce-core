// Package check provides a way to do health checks
package check

import (
	"github.com/pkg/errors"
)

// IsUnhealthyError checks if an error was caused by one of the erros in the checkers
func IsUnhealthyError(err error) bool {
	cause := errors.Cause(err)

	switch cause {
	case ErrCouldNotGetHost, ErrCouldNotPing,
		ErrCouldNotQuery, ErrCouldNotGetRows, ErrCouldNotScan:
		return true
	default:
		return false
	}
}

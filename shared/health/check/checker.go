// Package check provides a way to do health checks
package check

import (
	"context"
)

// Checker defines how to check.
type Checker interface {
	Check(ctx context.Context) error
}

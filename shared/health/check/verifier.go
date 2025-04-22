// Package check provides a way to do health checks
package check

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Verifier defines a way to verify the checks
type Verifier interface {
	Verify(ctx context.Context) error
}

type verifier struct {
	checkers []Checker
}

// NewVerifier adds the ability to create a default checkers implementation.
func NewVerifier(checkers []Checker) Verifier {
	return &verifier{checkers}
}

func (v *verifier) Verify(ctx context.Context) error {
	var group errgroup.Group

	for _, checker := range v.checkers {
		checker := checker

		group.Go(func() error {
			return checker.Check(ctx)
		})
	}

	return group.Wait()
}

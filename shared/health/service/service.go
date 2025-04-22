// Package service provides a way to do health checks
package service

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/shared/health/check"
)

// Service for health
type Service interface {
	Check(ctx context.Context) (*Response, error)
}

// NewService creates an service with all the checks.
func NewService(checkers []check.Checker) Service {
	verifier := check.NewVerifier(checkers)

	return &service{verifier}
}

type service struct {
	verifier check.Verifier
}

// swagger:route GET /health health HealthCheck
//
// Check if the service is healthy
//
//	Produces:
//	- application/json
//
//	Responses:
//	  200: HealthCheckResponse
func (s *service) Check(ctx context.Context) (*Response, error) {
	if err := s.verifier.Verify(ctx); err != nil {
		return nil, err
	}

	return &Response{"ok"}, nil
}

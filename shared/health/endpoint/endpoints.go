// Package endpoint contains health endpoints
package endpoint

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/shared/health/service"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints defined by health.
type Endpoints struct {
	CheckEndpoint endpoint.Endpoint
}

// New for health
func New(hs service.Service) *Endpoints {
	checkEndpoint := newCheckEndpoint(hs)

	return &Endpoints{CheckEndpoint: checkEndpoint}
}

func newCheckEndpoint(hs service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return hs.Check(ctx)
	}
}

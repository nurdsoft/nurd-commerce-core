package addressclient

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/internal/address/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/address/service"
)

type Client interface {
	GetAddress(ctx context.Context, req *entities.GetAddressRequest) (*entities.Address, error)
	GetDefaultAddress(ctx context.Context) (*entities.Address, error)
}

func NewClient(svc service.Service) Client {
	return &localClient{svc}
}

type localClient struct {
	svc service.Service
}

func (c *localClient) GetAddress(ctx context.Context, req *entities.GetAddressRequest) (*entities.Address, error) {
	return c.svc.GetAddress(ctx, req)
}

func (c *localClient) GetDefaultAddress(ctx context.Context) (*entities.Address, error) {
	return c.svc.GetDefaultAddress(ctx)
}

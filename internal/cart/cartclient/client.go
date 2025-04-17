package cartclient

import (
	"context"

	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/service"
)

type Client interface {
	GetCartItems(ctx context.Context) (*entities.GetCartItemsResponse, error)
	GetShippingRateByID(ctx context.Context, shippingRateID uuid.UUID) (*entities.CartShippingRate, error)
	GetCart(ctx context.Context) (*entities.Cart, error)
}

func NewClient(svc service.Service) Client {
	return &localClient{svc}
}

type localClient struct {
	svc service.Service
}

func (c *localClient) GetCartItems(ctx context.Context) (*entities.GetCartItemsResponse, error) {
	return c.svc.GetCartItems(ctx)
}

func (c *localClient) GetShippingRateByID(ctx context.Context, shippingRateID uuid.UUID) (*entities.CartShippingRate, error) {
	return c.svc.GetShippingRateByID(ctx, shippingRateID)
}

func (c *localClient) GetCart(ctx context.Context) (*entities.Cart, error) {
	return c.svc.GetCart(ctx)
}

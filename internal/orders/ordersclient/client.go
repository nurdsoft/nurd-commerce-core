package ordersclient

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/internal/orders/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/service"
)

type Client interface {
	ProcessPaymentSucceeded(ctx context.Context, paymentID string) error
	ProcessPaymentFailed(ctx context.Context, paymentID string) error
	ProcessOrderStatus(ctx context.Context, req *entities.UpdateOrderRequest) error
}

func NewClient(svc service.Service) Client {
	return &localClient{svc}
}

type localClient struct {
	svc service.Service
}

func (c *localClient) ProcessPaymentSucceeded(ctx context.Context, paymentID string) error {
	return c.svc.ProcessPaymentSucceeded(ctx, paymentID)
}

func (c *localClient) ProcessPaymentFailed(ctx context.Context, paymentID string) error {
	return c.svc.ProcessPaymentFailed(ctx, paymentID)
}

func (c *localClient) ProcessOrderStatus(ctx context.Context, req *entities.UpdateOrderRequest) error {
	return c.svc.UpdateOrder(ctx, req)
}

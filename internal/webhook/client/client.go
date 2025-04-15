package client

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/internal/webhook/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/webhook/service"
)

type Client interface {
	NotifyOrderStatusChange(ctx context.Context, req *entities.NotifyOrderStatusChangeRequest) error
}

func NewClient(svc service.Service) Client {
	return &localClient{svc}
}

type localClient struct {
	svc service.Service
}

func (c *localClient) NotifyOrderStatusChange(ctx context.Context, req *entities.NotifyOrderStatusChangeRequest) error {
	return c.svc.NotifyOrderStatusChange(ctx, req)
}

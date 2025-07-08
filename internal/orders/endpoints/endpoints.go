package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/service"
)

type Endpoints struct {
	CreateOrderEndpoint endpoint.Endpoint
	ListOrdersEndpoint  endpoint.Endpoint
	GetOrderEndpoint    endpoint.Endpoint
	CancelOrderEndpoint endpoint.Endpoint
	UpdateOrderEndpoint endpoint.Endpoint
	RefundOrderEndpoint endpoint.Endpoint
}

func New(svc service.Service) *Endpoints {
	return &Endpoints{
		CreateOrderEndpoint: makeCreateOrderEndpoint(svc),
		ListOrdersEndpoint:  makeListOrdersEndpoint(svc),
		GetOrderEndpoint:    makeGetOrderEndpoint(svc),
		CancelOrderEndpoint: makeCancelOrderEndpoint(svc),
		UpdateOrderEndpoint: makeUpdateOrderEndpoint(svc),
		RefundOrderEndpoint: makeRefundOrderEndpoint(svc),
	}
}

func makeCreateOrderEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.CreateOrderRequest)
		return svc.CreateOrder(ctx, req)
	}
}

func makeListOrdersEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.ListOrdersRequest)
		return svc.ListOrders(ctx, req)
	}
}

func makeGetOrderEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.GetOrderRequest)
		return svc.GetOrder(ctx, req)
	}
}

func makeCancelOrderEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.CancelOrderRequest)
		return nil, svc.CancelOrder(ctx, req)
	}
}

func makeUpdateOrderEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.UpdateOrderRequest)
		return nil, svc.UpdateOrder(ctx, req)
	}
}

func makeRefundOrderEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.RefundOrderRequest)
		return svc.RefundOrder(ctx, req)
	}
}

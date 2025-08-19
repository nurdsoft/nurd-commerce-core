package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/service"
)

type Endpoints struct {
	UpdateCartItemEndpoint          endpoint.Endpoint
	GetCartItemsEndpoint            endpoint.Endpoint
	RemoveCartItemEndpoint          endpoint.Endpoint
	ClearCartItemsEndpoint          endpoint.Endpoint
	GetShippingRateEndpoint         endpoint.Endpoint
	SetCartItemShippingRateEndpoint endpoint.Endpoint
	GetTaxRateEndpoint              endpoint.Endpoint
	CreateCartShippingRatesEndpoint endpoint.Endpoint
}

func New(svc service.Service) *Endpoints {
	return &Endpoints{
		UpdateCartItemEndpoint:          makeAddCartItem(svc),
		GetCartItemsEndpoint:            makeGetCartItems(svc),
		RemoveCartItemEndpoint:          makeRemoveCartItem(svc),
		ClearCartItemsEndpoint:          makeClearCartItems(svc),
		GetShippingRateEndpoint:         makeGetShippingRate(svc),
		SetCartItemShippingRateEndpoint: makeSetCartItemShippingRate(svc),
		GetTaxRateEndpoint:              makeGetTaxRate(svc),
		CreateCartShippingRatesEndpoint: makeCreateCartShippingRates(svc),
	}
}

func makeAddCartItem(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.UpdateCartItemRequest)
		return svc.UpdateCartItem(ctx, req)
	}
}

func makeGetCartItems(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		return svc.GetCartItems(ctx)
	}
}

func makeRemoveCartItem(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(string)
		return nil, svc.RemoveCartItem(ctx, req)
	}
}

func makeClearCartItems(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		return nil, svc.ClearCartItems(ctx)
	}
}

func makeGetTaxRate(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.GetTaxRateRequest)
		return svc.GetTaxRate(ctx, req)
	}
}

func makeGetShippingRate(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.GetShippingRateRequest)
		return svc.GetShippingRate(ctx, req)
	}
}

func makeCreateCartShippingRates(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.CreateCartShippingRatesRequest)
		return svc.CreateCartShippingRates(ctx, req)
	}
}

func makeSetCartItemShippingRate(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.SetCartItemShippingRateRequest)
		return nil, svc.SetCartItemShippingRate(ctx, req)
	}
}

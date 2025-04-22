package endpoints

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/internal/product/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/service"
	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	CreateProductEndpoint        endpoint.Endpoint
	GetProductEndpoint           endpoint.Endpoint
	CreateProductVariantEndpoint endpoint.Endpoint
	GetProductVariantEndpoint    endpoint.Endpoint
}

func New(svc service.Service) *Endpoints {
	return &Endpoints{
		CreateProductEndpoint:        makeCreateProduct(svc),
		GetProductEndpoint:           makeGetProduct(svc),
		CreateProductVariantEndpoint: makeCreateProductVariant(svc),
		GetProductVariantEndpoint:    makeGetProductVariant(svc),
	}
}

func makeCreateProduct(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.CreateProductRequest) //nolint:errcheck

		return svc.CreateProduct(ctx, req)
	}
}

func makeGetProduct(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.GetProductRequest) //nolint:errcheck

		return svc.GetProduct(ctx, req)
	}
}

func makeCreateProductVariant(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.CreateProductVariantRequest) //nolint:errcheck

		return svc.CreateProductVariant(ctx, req)
	}
}

func makeGetProductVariant(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.GetProductVariantRequest) //nolint:errcheck

		return svc.GetProductVariant(ctx, req)
	}
}

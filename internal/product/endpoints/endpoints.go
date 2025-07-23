package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/service"
)

type Endpoints struct {
	CreateProductEndpoint        endpoint.Endpoint
	GetProductEndpoint           endpoint.Endpoint
	CreateProductVariantEndpoint endpoint.Endpoint
	GetProductVariantEndpoint    endpoint.Endpoint
	ListProductVariantsEndpoint  endpoint.Endpoint
	ListProductsEndpoint         endpoint.Endpoint
}

func New(svc service.Service) *Endpoints {
	return &Endpoints{
		CreateProductEndpoint:        makeCreateProduct(svc),
		GetProductEndpoint:           makeGetProduct(svc),
		CreateProductVariantEndpoint: makeCreateProductVariant(svc),
		GetProductVariantEndpoint:    makeGetProductVariant(svc),
		ListProductVariantsEndpoint:  makeListProductVariants(svc),
		ListProductsEndpoint:         makeListProducts(svc),
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

func makeListProductVariants(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.ListProductVariantsRequest) //nolint:errcheck

		return svc.ListProductVariants(ctx, req)
	}
}

func makeListProducts(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.ListProductsRequest) //nolint:errcheck

		return svc.ListProducts(ctx, req)
	}
}

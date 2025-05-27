package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/service"
)

type Endpoints struct {
	AddToWishlistEndpoint                endpoint.Endpoint
	RemoveFromWishlistEndpoint           endpoint.Endpoint
	GetWishlistEndpoint                  endpoint.Endpoint
	GetMoreFromWishlistEndpoint          endpoint.Endpoint
	GetWishlistProductTimestampsEndpoint endpoint.Endpoint
}

func New(svc service.Service) *Endpoints {
	return &Endpoints{
		AddToWishlistEndpoint:                makeAddToWishlist(svc),
		RemoveFromWishlistEndpoint:           makeRemoveFromWishlist(svc),
		GetWishlistEndpoint:                  makeGetWishlist(svc),
		GetMoreFromWishlistEndpoint:          makeGetMoreFromWishlist(svc),
		GetWishlistProductTimestampsEndpoint: makeGetWishlistProductTimestamps(svc),
	}
}

func makeAddToWishlist(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.AddToWishlistRequest) //nolint:errcheck
		err := svc.AddToWishlist(ctx, req)

		return nil, err
	}
}

func makeRemoveFromWishlist(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.RemoveFromWishlistRequest) //nolint:errcheck
		err := svc.RemoveFromWishlist(ctx, req)

		return nil, err
	}
}

func makeGetWishlist(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.GetWishlistRequest) //nolint:errcheck
		wishlist, err := svc.GetWishlist(ctx, req)

		return wishlist, err
	}
}

func makeGetMoreFromWishlist(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.GetMoreFromWishlistRequest) //nolint:errcheck
		return svc.GetMoreFromWishlist(ctx, req)
	}
}

func makeGetWishlistProductTimestamps(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*entities.GetWishlistProductTimestampsRequest) //nolint:errcheck
		timestamps, err := svc.GetWishlistProductTimestamps(ctx, req)

		return timestamps, err
	}
}

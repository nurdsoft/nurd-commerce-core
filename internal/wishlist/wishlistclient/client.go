package wishlistclient

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/service"
)

type Client interface {
	BulkRemoveFromWishlist(ctx context.Context, req *entities.BulkRemoveFromWishlistRequest) error
}

func NewClient(svc service.Service) Client {
	return &localClient{svc}
}

type localClient struct {
	svc service.Service
}

func (c *localClient) BulkRemoveFromWishlist(ctx context.Context, req *entities.BulkRemoveFromWishlistRequest) error {
	return c.svc.BulkRemoveFromWishlist(ctx, req)
}

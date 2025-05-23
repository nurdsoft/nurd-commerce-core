package productclient

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/internal/product/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/service"
)

type Client interface {
	CreateProduct(ctx context.Context, request *entities.CreateProductRequest) (*entities.Product, error)
	GetProduct(ctx context.Context, request *entities.GetProductRequest) (*entities.Product, error)
	GetProductsByIDs(ctx context.Context, ids []string) ([]entities.Product, error)
	UpdateProduct(ctx context.Context, request *entities.UpdateProductRequest) error
	CreateProductVariant(ctx context.Context, req *entities.CreateProductVariantRequest) (*entities.ProductVariant, error)
	GetProductVariant(ctx context.Context, req *entities.GetProductVariantRequest) (*entities.ProductVariant, error)
	GetProductVariantByID(ctx context.Context, variantID string) (*entities.ProductVariant, error)
}

func NewClient(svc service.Service) Client {
	return &localClient{svc}
}

type localClient struct {
	svc service.Service
}

func (c *localClient) CreateProduct(ctx context.Context, req *entities.CreateProductRequest) (*entities.Product, error) {
	return c.svc.CreateProduct(ctx, req)
}

func (c *localClient) GetProduct(ctx context.Context, req *entities.GetProductRequest) (*entities.Product, error) {
	return c.svc.GetProduct(ctx, req)
}

func (c *localClient) UpdateProduct(ctx context.Context, req *entities.UpdateProductRequest) error {
	return c.svc.UpdateProduct(ctx, req)
}

func (c *localClient) CreateProductVariant(ctx context.Context, req *entities.CreateProductVariantRequest) (*entities.ProductVariant, error) {
	return c.svc.CreateProductVariant(ctx, req)
}

func (c *localClient) GetProductVariant(ctx context.Context, req *entities.GetProductVariantRequest) (*entities.ProductVariant, error) {
	return c.svc.GetProductVariant(ctx, req)
}

func (c *localClient) GetProductVariantByID(ctx context.Context, variantID string) (*entities.ProductVariant, error) {
	return c.svc.GetProductVariantByID(ctx, variantID)
}

func (c *localClient) GetProductsByIDs(ctx context.Context, ids []string) ([]entities.Product, error) {
	return c.svc.GetProductsByIDs(ctx, ids)
}

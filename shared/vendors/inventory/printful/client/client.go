package client

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strconv"

	inventoryEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/printful/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/printful/service"
)

type Client interface {
	GetSyncProducts(ctx context.Context, req inventoryEntities.ListProductsRequest) (inventoryEntities.ListProductsResponse, error)
	GetSyncProduct(ctx context.Context, id string) (*inventoryEntities.Product, error)
}

type client struct {
	service service.Service
}

func NewClient(service service.Service) Client {
	return &client{
		service: service,
	}
}

func (c *client) GetSyncProducts(ctx context.Context, req inventoryEntities.ListProductsRequest) (inventoryEntities.ListProductsResponse, error) {
	response, err := c.service.GetSyncProducts(ctx, entities.GetSyncProductsRequest{
		Search: req.Search,
		Limit:  req.PageSize,
		Offset: (req.Page - 1) * req.PageSize,
	})
	if err != nil {
		return inventoryEntities.ListProductsResponse{}, err
	}
	products := make([]inventoryEntities.ProductResponse, len(response.Result))
	for i, p := range response.Result {
		products[i] = inventoryEntities.ProductResponse{
			ID:       strconv.Itoa(p.ID),
			Name:     p.Name,
			ImageURL: &p.ThumbnailURL,
		}
	}
	return inventoryEntities.ListProductsResponse{
		Data: products,
		Pagination: inventoryEntities.PaginationMeta{
			Page:       req.Page,
			PageSize:   req.PageSize,
			Total:      response.Paging.Total,
			TotalPages: int(math.Ceil(float64(response.Paging.Total) / float64(req.PageSize))),
		},
	}, nil
}

func (c *client) GetSyncProduct(ctx context.Context, id string) (*inventoryEntities.Product, error) {
	productID, err := strconv.Atoi(id)
	if err != nil {
		return nil, errors.New("invalid product ID")
	}
	response, err := c.service.GetSyncProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	variants := make([]inventoryEntities.ProductVariant, len(response.Result.SyncVariants))
	for i, v := range response.Result.SyncVariants {
		attributes := make(map[string]string)
		if v.Size != "" {
			attributes["size"] = v.Size
		}
		if v.Color != "" {
			attributes["color"] = v.Color
		}
		attributesJSON, err := json.Marshal(attributes)
		if err != nil {
			return nil, err
		}

		variants[i] = inventoryEntities.ProductVariant{
			ID:         strconv.Itoa(v.ID),
			Name:       v.Name,
			SKU:        v.SKU,
			ImageURL:   &v.Files[1].PreviewURL,
			Attributes: (*json.RawMessage)(&attributesJSON),
		}
	}

	return &inventoryEntities.Product{
		ID:       strconv.Itoa(response.Result.SyncProduct.ID),
		Name:     response.Result.SyncProduct.Name,
		ImageURL: &response.Result.SyncProduct.ThumbnailURL,
		Variants: variants,
	}, nil
}

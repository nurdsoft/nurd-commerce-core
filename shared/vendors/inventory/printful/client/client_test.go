package client

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	inventoryEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/entities"
	printfulEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/printful/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/printful/service"
	"github.com/stretchr/testify/assert"
)

func TestClient_GetSyncProducts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockResp := &printfulEntities.SyncProductsResponse{
			Paging: printfulEntities.SyncProductsPaging{Total: 1, Limit: 10, Offset: 0},
			Result: []printfulEntities.SyncProduct{{ID: 1, Name: "Test Product", ThumbnailURL: "http://img"}},
		}
		mockService.EXPECT().GetSyncProducts(ctx, gomock.Any()).Return(mockResp, nil)

		resp, err := client.GetSyncProducts(ctx, inventoryEntities.ListProductsRequest{Page: 1, PageSize: 10})
		assert.NoError(t, err)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "Test Product", resp.Data[0].Name)
	})

	t.Run("error from service", func(t *testing.T) {
		mockService.EXPECT().GetSyncProducts(ctx, gomock.Any()).Return(nil, errors.New("service error"))
		resp, err := client.GetSyncProducts(ctx, inventoryEntities.ListProductsRequest{Page: 1, PageSize: 10})
		assert.Error(t, err)
		assert.Empty(t, resp.Data)
	})
}

func TestClient_GetSyncProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	client := NewClient(mockService)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockResp := &printfulEntities.GetSyncProductResponse{
			Result: printfulEntities.GetSyncProductResult{
				SyncProduct:  printfulEntities.SyncProduct{ID: 2, Name: "Product2", ThumbnailURL: "http://img2"},
				SyncVariants: []printfulEntities.SyncVariant{{ID: 3, Name: "Var1", SKU: "SKU1", Files: []printfulEntities.SyncVariantFile{{}, {PreviewURL: "http://varimg"}}}},
			},
		}
		mockService.EXPECT().GetSyncProduct(ctx, 2).Return(mockResp, nil)
		prod, err := client.GetSyncProduct(ctx, "2")
		assert.NoError(t, err)
		assert.Equal(t, "Product2", prod.Name)
		assert.Len(t, prod.Variants, 1)
		assert.Equal(t, "Var1", prod.Variants[0].Name)
		assert.Equal(t, "SKU1", prod.Variants[0].SKU)
		assert.Equal(t, "http://varimg", *prod.Variants[0].ImageURL)
	})

	t.Run("invalid id", func(t *testing.T) {
		prod, err := client.GetSyncProduct(ctx, "notanint")
		assert.Error(t, err)
		assert.Nil(t, prod)
	})

	t.Run("error from service", func(t *testing.T) {
		mockService.EXPECT().GetSyncProduct(ctx, 5).Return(nil, errors.New("service error"))
		prod, err := client.GetSyncProduct(ctx, "5")
		assert.Error(t, err)
		assert.Nil(t, prod)
	})
}

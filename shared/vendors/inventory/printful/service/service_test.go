package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	printfulConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/printful/config"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/printful/entities"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestService_GetSyncProducts(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/sync/products" {
				t.Errorf("Expected path /sync/products, got %s", r.URL.Path)
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer test-token" {
				t.Errorf("Expected Authorization header 'Bearer test-token', got %s", authHeader)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"code": 200,
				"result": [
					{
						"id": 386905543,
						"external_id": "68754481a46ea5",
						"name": "Unisex classic tee",
						"variants": 20,
						"synced": 20,
						"thumbnail_url": "https://files.cdn.printful.com/files/50d/50d5ce8a374179c962098208ee11abe5_preview.png",
						"is_ignored": false
					}
				],
				"extra": [],
				"paging": {
					"total": 1,
					"limit": 20,
					"offset": 0
				}
			}`))
		}))
		defer server.Close()

		config := printfulConfig.Config{
			OAuthToken: "test-token",
			BaseURL:    server.URL,
		}

		svc := New(config, &http.Client{}, zap.NewNop().Sugar())

		response, err := svc.GetSyncProducts(ctx, entities.GetSyncProductsRequest{})


		assert.NoError(t, err)	
		assert.NotEmpty(t, response)
		assert.Equal(t, 200, response.Code)
		assert.Len(t, response.Result, 1)

		product := response.Result[0]
		assert.Equal(t, 386905543, product.ID)
		assert.Equal(t, "Unisex classic tee", product.Name)
		assert.Equal(t, "68754481a46ea5", product.ExternalID)
	})

	t.Run("empty response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"code": 200,
				"result": [],
				"extra": [],
				"paging": {
					"total": 0,
					"limit": 20,
					"offset": 0
				}
			}`))
		}))
		defer server.Close()

		config := printfulConfig.Config{
			OAuthToken: "test-token",
			BaseURL:    server.URL,
		}

		svc := New(config, &http.Client{}, zap.NewNop().Sugar())

		response, err := svc.GetSyncProducts(ctx, entities.GetSyncProductsRequest{})

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.Code)
		assert.Empty(t, response.Result)
	})

	t.Run("unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{
				"code": 401,
				"result": "The access token provided is invalid.",
				"error": {
					"reason": "Unauthorized",
					"message": "The access token provided is invalid."
				}
			}`))
		}))
		defer server.Close()

		config := printfulConfig.Config{
			OAuthToken: "test-token",
			BaseURL:    server.URL,
		}

		svc := New(config, &http.Client{}, zap.NewNop().Sugar())

		response, err := svc.GetSyncProducts(ctx, entities.GetSyncProductsRequest{})

		assert.Error(t, err)
		assert.Empty(t, response)
	})
}

func TestService_GetSyncProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/sync/products/13" {
				t.Errorf("Expected path /sync/products/13, got %s", r.URL.Path)
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer test-token" {
				t.Errorf("Expected Authorization header 'Bearer test-token', got %s", authHeader)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"code": 200,
				"result": {
					"sync_product": {
						"id": 13,
						"external_id": "4235234213",
						"name": "T-shirt",
						"variants": 10,
						"synced": 10,
						"thumbnail_url": "https://your-domain.com/path/to/image.png",
						"is_ignored": true
					},
					"sync_variants": [
						{
							"id": 10,
							"external_id": "12312414",
							"sync_product_id": 71,
							"name": "Red T-Shirt",
							"synced": true,
							"variant_id": 3001,
							"retail_price": "29.99",
							"currency": "USD",
							"is_ignored": true,
							"sku": "SKU1234",
							"product": {
								"variant_id": 3001,
								"product_id": 301,
								"image": "https://files.cdn.printful.com/products/71/5309_1581412541.jpg",
								"name": "Bella + Canvas 3001 Unisex Short Sleeve Jersey T-Shirt with Tear Away Label (White / 4XL)"
							},
							"files": [
								{
									"type": "default",
									"id": 10,
									"url": "https://www.example.com/files/tshirts/example.png",
									"options": [
										{
											"id": "template_type",
											"value": "native"
										}
									],
									"hash": "ea44330b887dfec278dbc4626a759547",
									"filename": "shirt1.png",
									"mime_type": "image/png",
									"size": 45582633,
									"width": 1000,
									"height": 1000,
									"dpi": 300,
									"status": "ok",
									"created": 1590051937,
									"thumbnail_url": "https://files.cdn.printful.com/files/ea4/ea44330b887dfec278dbc4626a759547_thumb.png",
									"preview_url": "https://files.cdn.printful.com/files/ea4/ea44330b887dfec278dbc4626a759547_thumb.png",
									"visible": true,
									"is_temporary": false,
									"stitch_count_tier": "stitch_tier_1"
								}
							],
							"options": [
								{
									"id": "embroidery_type",
									"value": "flat"
								}
							],
							"main_category_id": 24,
							"warehouse_product_id": 3002,
							"warehouse_product_variant_id": 3002,
							"size": "XS",
							"color": "White",
							"availability_status": "active"
						}
					]
				}
			}`))
		}))
		defer server.Close()

		config := printfulConfig.Config{
			OAuthToken: "test-token",
			BaseURL:    server.URL,
		}

		svc := New(config, &http.Client{}, zap.NewNop().Sugar())

		response, err := svc.GetSyncProduct(ctx, 13)

		assert.NoError(t, err)
		assert.NotEmpty(t, response)
		assert.Equal(t, 200, response.Code)
		assert.Equal(t, 13, response.Result.SyncProduct.ID)
		assert.Equal(t, "T-shirt", response.Result.SyncProduct.Name)
		assert.Equal(t, "4235234213", response.Result.SyncProduct.ExternalID)
		assert.Len(t, response.Result.SyncVariants, 1)

		variant := response.Result.SyncVariants[0]
		assert.Equal(t, 10, variant.ID)
		assert.Equal(t, "Red T-Shirt", variant.Name)
		assert.Equal(t, "SKU1234", variant.SKU)
		assert.Equal(t, "29.99", variant.RetailPrice)
		assert.Equal(t, "USD", variant.Currency)
		assert.Equal(t, "active", variant.AvailabilityStatus)
	})
} 
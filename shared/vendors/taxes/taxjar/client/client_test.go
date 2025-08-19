package client

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/taxjar/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/taxjar/taxjar-go"
)

func TestCalculateTax(t *testing.T) {
	ctx := context.Background()
	t.Run("should calculate tax", func(t *testing.T) {
		mockSvc := service.NewMockService(gomock.NewController(t))
		client := NewClient(mockSvc)

		req := &entities.CalculateTaxRequest{
			FromAddress: entities.Address{
				Street:     "123 Main St",
				City:       "Anytown",
				State:      "CA",
				PostalCode: "12345",
				Country:    "US",
			},
			ToAddress: entities.Address{
				Street:     "123 Main St",
				City:       "Anytown",
				State:      "CA",
				PostalCode: "12345",
				Country:    "US",
			},
			ShippingAmount: decimal.NewFromInt(10),
			TaxItems: []entities.TaxItem{
				{
					Price:     decimal.NewFromInt(49),
					Quantity:  1,
					Reference: "sku-1",
					TaxCode:   "20010",
				},
				{
					Price:     decimal.NewFromInt(34),
					Quantity:  2,
					Reference: "sku-2",
					TaxCode:   "20010",
				},
			},
		}

		mockSvc.EXPECT().
			CalculateTax(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, params taxjar.TaxForOrderParams) (*taxjar.TaxForOrderResponse, error) {
				assert.Equal(t, "US", params.FromCountry)
				assert.Equal(t, "12345", params.FromZip)
				assert.Equal(t, "CA", params.FromState)
				assert.Equal(t, "Anytown", params.FromCity)
				assert.Equal(t, "123 Main St", params.FromStreet)
				assert.Equal(t, "US", params.ToCountry)
				assert.Equal(t, "12345", params.ToZip)
				assert.Equal(t, "CA", params.ToState)
				assert.Equal(t, "Anytown", params.ToCity)
				assert.Equal(t, "123 Main St", params.ToStreet)

				assert.Equal(t, 10.00, params.Shipping)

				assert.Equal(t, 2, len(params.LineItems))
				assert.Equal(t, "sku-1", params.LineItems[0].ID)
				assert.Equal(t, 1, params.LineItems[0].Quantity)
				assert.Equal(t, "20010", params.LineItems[0].ProductTaxCode)
				assert.Equal(t, 49.00, params.LineItems[0].UnitPrice)
				assert.Equal(t, "sku-2", params.LineItems[1].ID)
				assert.Equal(t, 2, params.LineItems[1].Quantity)
				assert.Equal(t, "20010", params.LineItems[1].ProductTaxCode)
				assert.Equal(t, 34.00, params.LineItems[1].UnitPrice)

				return &taxjar.TaxForOrderResponse{
					Tax: taxjar.Tax{
						AmountToCollect:  12.70,
						OrderTotalAmount: 127.00,
						Rate:             0.1,
						Shipping:         10,
					},
				}, nil
			})

		res, err := client.CalculateTax(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.True(t, res.Tax.Equal(decimal.NewFromFloat(12.70)))
		assert.True(t, res.TotalAmount.Equal(decimal.NewFromFloat(127.00)))
		assert.Equal(t, "USD", res.Currency)
	})

	t.Run("should return error if service returns error", func(t *testing.T) {
		mockSvc := service.NewMockService(gomock.NewController(t))
		client := NewClient(mockSvc)

		mockSvc.EXPECT().CalculateTax(ctx, gomock.Any()).Return(nil, errors.New("error"))

		req := &entities.CalculateTaxRequest{
			ToAddress: entities.Address{
				Street:     "123 Main St",
				City:       "Anytown",
				State:      "CA",
				PostalCode: "12345",
				Country:    "US",
			},
			ShippingAmount: decimal.NewFromInt(10),
			TaxItems: []entities.TaxItem{
				{
					Price:     decimal.NewFromInt(49),
					Quantity:  1,
					Reference: "sku-1",
					TaxCode:   "20010",
				},
			},
		}

		_, err := client.CalculateTax(ctx, req)
		assert.Error(t, err)
	})
}

package client

import (
	"context"
	"encoding/json"

	commerceJson "github.com/nurdsoft/nurd-commerce-core/shared/json"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/taxjar/service"
	"github.com/shopspring/decimal"
	"github.com/taxjar/taxjar-go"
)

type Client interface {
	CalculateTax(ctx context.Context, req *entities.CalculateTaxRequest) (*entities.CalculateTaxResponse, error)
}

func NewClient(svc service.Service) Client {
	return &localClient{svc}
}

type localClient struct {
	svc service.Service
}

func (c *localClient) CalculateTax(ctx context.Context, req *entities.CalculateTaxRequest) (*entities.CalculateTaxResponse, error) {
	taxItems := make([]taxjar.TaxLineItem, len(req.TaxItems))
	for i, item := range req.TaxItems {
		if item.TaxCode == "" {
			item.TaxCode = "20010"
		}

		taxItems[i] = taxjar.TaxLineItem{
			ID:             item.Reference,
			Quantity:       item.Quantity,
			ProductTaxCode: item.TaxCode,
			UnitPrice:      item.Price.InexactFloat64(),
		}
	}

	taxParams := taxjar.TaxForOrderParams{
		ToCountry: req.ToAddress.Country,
		ToZip:     req.ToAddress.PostalCode,
		ToState:   req.ToAddress.State,
		ToCity:    req.ToAddress.City,
		ToStreet:  req.ToAddress.Street,

		Shipping: req.ShippingAmount.InexactFloat64(),

		LineItems: taxItems,
	}

	if req.FromAddress != nil {
		taxParams.FromCountry = req.FromAddress.Country
		taxParams.FromZip = req.FromAddress.PostalCode
		taxParams.FromState = req.FromAddress.State
		taxParams.FromCity = req.FromAddress.City
		taxParams.FromStreet = req.FromAddress.Street
	}

	tax, err := c.svc.CalculateTax(ctx, taxParams)
	if err != nil {
		return nil, err
	}

	taxBreakdownJSON, err := json.Marshal(tax.Tax.Breakdown)
	if err != nil {
		return nil, err
	}

	return &entities.CalculateTaxResponse{
		Tax:         decimal.NewFromFloat(tax.Tax.AmountToCollect),
		TotalAmount: decimal.NewFromFloat(tax.Tax.OrderTotalAmount),
		Currency:    "USD",
		Breakdown:   commerceJson.JSON(taxBreakdownJSON),
	}, nil
}

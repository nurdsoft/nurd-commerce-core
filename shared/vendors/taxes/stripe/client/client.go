package client

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/entities"
	stripeEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/service"
	"github.com/shopspring/decimal"
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
	stripeReq := &stripeEntities.CalculateTaxRequest{
		ShippingAmount: req.ShippingAmount,
		ToAddress: stripeEntities.Address{
			Line1:      req.ToAddress.Street,
			City:       req.ToAddress.City,
			State:      req.ToAddress.State,
			PostalCode: req.ToAddress.PostalCode,
			Country:    req.ToAddress.Country,
		},
		TaxItems: mapStripeTaxItems(req.TaxItems),
	}

	if req.FromAddress != nil {
		stripeReq.FromAddress = &stripeEntities.Address{
			Line1:      req.FromAddress.Street,
			City:       req.FromAddress.City,
			State:      req.FromAddress.State,
			PostalCode: req.FromAddress.PostalCode,
			Country:    req.FromAddress.Country,
		}
	}

	res, err := c.svc.CalculateTax(ctx, stripeReq)
	if err != nil {
		return nil, err
	}

	// convert the tax rate from minor units back to major units for human readability
	taxInMajorUnits := res.Tax.Div(decimal.NewFromInt(100))
	totalAmountInMajorUnits := res.TotalAmount.Div(decimal.NewFromInt(100))

	return &entities.CalculateTaxResponse{
		Tax:         taxInMajorUnits,
		TotalAmount: totalAmountInMajorUnits,
		Currency:    res.Currency,
		Breakdown:   res.Breakdown,
	}, nil
}

func mapStripeTaxItems(items []entities.TaxItem) []stripeEntities.TaxItem {
	stripeItems := make([]stripeEntities.TaxItem, len(items))
	for i, item := range items {
		stripeItems[i] = stripeEntities.TaxItem{
			// Stripe requires to provide the amount of the product with the no.of pieces being bought
			Price:     item.Price.Mul(decimal.NewFromInt(int64(item.Quantity))),
			Quantity:  item.Quantity,
			Reference: item.Reference,
			TaxCode:   item.TaxCode,
		}
	}

	return stripeItems
}

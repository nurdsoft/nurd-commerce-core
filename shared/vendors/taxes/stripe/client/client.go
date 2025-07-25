package client

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/entities"
	stripeEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/service"
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
		FromAddress: stripeEntities.Address{
			Line1:      req.FromAddress.Line1,
			City:       req.FromAddress.City,
			State:      req.FromAddress.State,
			PostalCode: req.FromAddress.PostalCode,
			Country:    req.FromAddress.Country,
		},
		ToAddress: stripeEntities.Address{
			Line1:      req.ToAddress.Line1,
			City:       req.ToAddress.City,
			State:      req.ToAddress.State,
			PostalCode: req.ToAddress.PostalCode,
			Country:    req.ToAddress.Country,
		},
		TaxItems: mapStripeTaxItems(req.TaxItems),
	}

	res, err := c.svc.CalculateTax(ctx, stripeReq)
	if err != nil {
		return nil, err
	}

	return &entities.CalculateTaxResponse{
		Tax:         res.Tax,
		TotalAmount: res.TotalAmount,
		Currency:    res.Currency,
		Breakdown:   res.Breakdown,
	}, nil
}

func mapStripeTaxItems(items []entities.TaxItem) []stripeEntities.TaxItem {
	stripeItems := make([]stripeEntities.TaxItem, len(items))
	for i, item := range items {
		stripeItems[i] = stripeEntities.TaxItem{
			Price:     item.Price,
			Quantity:  item.Quantity,
			Reference: item.Reference,
			TaxCode:   item.TaxCode,
		}
	}

	return stripeItems
}

package fakeprovider

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/shared/json"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/entities"
	"github.com/shopspring/decimal"
)

type Client interface {
	CalculateTax(ctx context.Context, req *entities.CalculateTaxRequest) (*entities.CalculateTaxResponse, error)
}

func NewClient() Client {
	return &localClient{}
}

type localClient struct{}

func (c *localClient) CalculateTax(ctx context.Context, req *entities.CalculateTaxRequest) (*entities.CalculateTaxResponse, error) {
	return &entities.CalculateTaxResponse{
		Tax:         decimal.NewFromFloat(3.99),
		TotalAmount: decimal.NewFromFloat(103.99),
		Currency:    "USD",
		Breakdown:   json.JSON(`{"tax": 3.99}`),
	}, nil
}

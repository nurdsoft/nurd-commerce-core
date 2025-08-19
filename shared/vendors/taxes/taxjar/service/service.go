package service

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/taxjar/config"
	"github.com/taxjar/taxjar-go"
)

type Service interface {
	CalculateTax(ctx context.Context, params taxjar.TaxForOrderParams) (*taxjar.TaxForOrderResponse, error)
}

func New(taxjarConfig config.Config) (Service, error) {
	taxjarClient := taxjar.NewClient(taxjar.Config{
		APIKey: taxjarConfig.Key,
		APIURL: taxjarConfig.URL,
	})

	return &service{&taxjarClient}, nil
}

type service struct {
	taxjarClient *taxjar.Config
}

func (s *service) CalculateTax(ctx context.Context, params taxjar.TaxForOrderParams) (*taxjar.TaxForOrderResponse, error) {
	return s.taxjarClient.TaxForOrder(params)
}

package service

import (
	"context"
	"fmt"
	"github.com/nurdsoft/nurd-commerce-core/internal/transport/http/client"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/entities"
	upsConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/ups/config"
	"go.uber.org/zap"
	"net/http"
)

type Service interface {
	ValidateAddress(ctx context.Context, address entities.Address) error
	GetShippingRates(ctx context.Context, shipment entities.Shipment) ([]entities.ShippingRate, error)
}

func New(httpClient *http.Client, config upsConfig.Config, logger *zap.SugaredLogger) (Service, error) {
	hc := client.New(fmt.Sprintf("https://%s", config.APIHost), httpClient, client.WithExternalCall(true))

	return &service{hc, config, logger}, nil
}

type service struct {
	httpClient client.Client
	config     upsConfig.Config
	logger     *zap.SugaredLogger
}

// GetShippingRates returns the estimated rates for the given shipping address and dimensions
// https://shipengine.github.io/shipengine-openapi/#operation/estimate_rates
func (s *service) GetShippingRates(ctx context.Context, shipment entities.Shipment) ([]entities.ShippingRate, error) {

	return nil, nil
}

// ValidateAddress return validation result for the given shipping address
// https://shipengine.github.io/shipengine-openapi/#operation/estimate_rates
func (s *service) ValidateAddress(ctx context.Context, address entities.Address) error {
	return nil
}

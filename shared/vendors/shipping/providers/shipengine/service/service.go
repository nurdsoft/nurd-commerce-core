package service

import (
	"context"
	"encoding/json"
	"fmt"
	shipengineConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/shipengine/config"
	"github.com/shopspring/decimal"
	"net/http"
	"strings"

	"github.com/nurdsoft/nurd-commerce-core/internal/transport/http/client"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/entities"
	shipengineEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/shipengine/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/shipengine/errors"
	shipengineErrors "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/shipengine/errors"
	"go.uber.org/zap"
)

type Service interface {
	ValidateAddress(ctx context.Context, address entities.Address) error
	GetShippingRates(ctx context.Context, shipment entities.Shipment) ([]entities.ShippingRate, error)
}

func New(httpClient *http.Client, config shipengineConfig.Config, logger *zap.SugaredLogger) (Service, error) {
	hc := client.New(fmt.Sprintf("https://%s", config.Host), httpClient, client.WithExternalCall(true))

	return &service{hc, config, logger}, nil
}

type service struct {
	httpClient client.Client
	config     shipengineConfig.Config
	logger     *zap.SugaredLogger
}

// GetShippingRates returns the estimated rates for the given shipping address and dimensions
// https://shipengine.github.io/shipengine-openapi/#operation/estimate_rates
func (s *service) GetShippingRates(ctx context.Context, shipment entities.Shipment) ([]entities.ShippingRate, error) {

	// Each Token is associated with different carriers Ids. Use the list GET /v1/carriers endpoint to get the list of carriers
	// Make sure to change the carrierIds on per environment basis
	carriers := strings.Split(s.config.CarrierIds, ",")
	if len(carriers) == 0 {
		return nil, moduleErrors.NewAPIError("SHIPENGINE_MISSING_CARRIERS")
	}

	req := shipengineEntities.ShippingRateRequest{
		CarrierIds:        carriers,
		FromCountryCode:   shipment.Origin.CountryCode,
		FromPostalCode:    shipment.Origin.PostalCode,
		FromCityLocality:  shipment.Origin.City,
		FromStateProvince: shipment.Origin.StateCode,
		ToCountryCode:     shipment.Destination.CountryCode,
		ToPostalCode:      shipment.Destination.PostalCode,
		ToCityLocality:    shipment.Destination.City,
		ToStateProvince:   shipment.Destination.StateCode,
		Weight: shipengineEntities.Weight{
			Value: shipment.Dimensions.Weight.InexactFloat64(),
			Unit:  "pound",
		},
		Dimensions: shipengineEntities.ObjectDimensions{
			Length: shipment.Dimensions.Length.InexactFloat64(),
			Width:  shipment.Dimensions.Width.InexactFloat64(),
			Height: shipment.Dimensions.Height.InexactFloat64(),
			Unit:   "inch",
		},
	}

	var res []shipengineEntities.EstimateRatesResponse

	err := s.httpClient.Post(ctx, s.getShipengineReqUrl("rates/estimate"), s.getShipengineApiHeaders(), req, &res)
	if err != nil {
		e := client.InvalidResponseError(err)
		if e != nil {
			var shipError shipengineErrors.ErrorObject
			_ = json.Unmarshal([]byte(e.Description), &shipError)

			if len(shipError.Errors) > 0 {
				for _, v := range shipError.Errors {
					if v.ErrorSource == "shipengine" {
						switch v.Message {
						case shipengineErrors.ErrInvalidToPostalCode, shipengineErrors.ErrEmptyToCountryCode:
							return nil, moduleErrors.NewAPIError("SHIPENGINE_INVALID_POSTAL_CODE")
						case shipengineErrors.ErrEmptyFromPostalCode:
							return nil, moduleErrors.NewAPIError("SHIPENGINE_INVALID_ORIGIN_POSTAL_CODE")
						case shipengineErrors.ErrEmptyCarrierId, shipengineErrors.ErrEmptyCarrierIds:
							return nil, moduleErrors.NewAPIError("SHIPENGINE_MISSING_CARRIERS")
						}
					}
				}
			}
		}
		return nil, moduleErrors.NewAPIError("SHIPENGINE_ERROR_GETTING_SHIPPING_RATES")
	}

	// Check for any error messages in the response from individual carriers
	// Note: Shipengine doesn't have a consistent place to put error messages in the response
	// TODO commenting this for now, even if one the carriers has an error, we still want to show the rates from other carriers
	/* if len(res) > 0 {
		for _, v := range res {
			log.Println("juan1", v.ErrorMessages)
			log.Println("juan2", v.ShippingAmount.Amount)
			if len(v.ErrorMessages) > 0 && v.ShippingAmount.Amount == 0 {
				switch v.ErrorMessages[0] {
				// only handle invalid country code error
				case shipengineErrors.CarrierErrUPSMissingDestination:
					return nil, &appErrors.ErrBadRequest{Message: "Invalid delivery address country code"}
				case shipengineErrors.CarrierErrUPSMaxWeight:
					return nil, &appErrors.ErrBadRequest{Message: "Package weight exceeds carrier limit"}
				default:
					return nil, &appErrors.ErrBadRequest{Message: fmt.Sprintf("Carrier error: %s", v.ErrorMessages[0])}
				}
			}
		}
	}*/

	allRates := make(map[string][]shipengineEntities.EstimateRatesResponse)
	bestRatesPerCarrier := make(map[string]shipengineEntities.EstimateRatesResponse)

	// group rates by carried id. Looks like "se-1140635"
	for _, estimate := range res {
		if estimate.EstimatedDeliveryDate.Valid {
			allRates[estimate.CarrierID] = append(allRates[estimate.CarrierID], estimate)
		}
	}

	// for each carrier, find the best rate i.e lowest shipping amount
	for carrierID, rates := range allRates {
		var bestRate shipengineEntities.EstimateRatesResponse

		for _, rate := range rates {
			if rate.ShippingAmount.Amount > 0 {
				if bestRate.ShippingAmount.Amount == 0 || rate.ShippingAmount.Amount < bestRate.ShippingAmount.Amount {
					bestRate = rate
				}
			}
		}
		bestRatesPerCarrier[carrierID] = bestRate
	}

	var shippingRates []entities.ShippingRate

	for _, bestRate := range bestRatesPerCarrier {
		shippingRates = append(shippingRates, entities.ShippingRate{
			Amount:                decimal.NewFromFloat(bestRate.ShippingAmount.Amount),
			Currency:              bestRate.ShippingAmount.Currency,
			CarrierName:           bestRate.CarrierFriendlyName,
			CarrierCode:           bestRate.CarrierCode,
			ServiceType:           bestRate.ServiceType,
			ServiceCode:           bestRate.ServiceCode,
			EstimatedDeliveryDate: bestRate.EstimatedDeliveryDate.Time,
		})
	}

	return shippingRates, nil
}

// ValidateAddress return validation result for the given shipping address
// https://shipengine.github.io/shipengine-openapi/#operation/estimate_rates
func (s *service) ValidateAddress(ctx context.Context, address entities.Address) error {
	// TODO Change to the actual validation address API once paid tier is enabled to test
	_, err := s.GetShippingRates(ctx,
		entities.Shipment{
			Origin: entities.Address{
				City:        "La Vergne",
				StateCode:   "TN",
				PostalCode:  "37086",
				CountryCode: "US",
			},
			Destination: address,
			Dimensions: entities.Dimensions{
				Length: decimal.NewFromFloat(1),
				Width:  decimal.NewFromFloat(1),
				Height: decimal.NewFromFloat(1),
				Weight: decimal.NewFromFloat(1),
			},
		})
	return err
}

func (s *service) getShipengineApiHeaders() map[string]string {
	return map[string]string{
		"API-Key":      s.config.Token,
		"Content-Type": "application/json",
	}
}

func (s *service) getShipengineReqUrl(path string) string {
	return fmt.Sprintf("/v1/%s", path)
}

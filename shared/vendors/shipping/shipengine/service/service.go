package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/nurdsoft/nurd-commerce-core/internal/transport/http/client"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/shipengine/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/shipengine/errors"
	shipengineErrors "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/shipengine/errors"
	"go.uber.org/zap"
)

type Service interface {
	GetRatesEstimate(ctx context.Context, from, to entities.ShippingAddress, dimensions entities.Dimensions) ([]entities.EstimateRatesResponse, error)
}

func New(httpClient *http.Client, config shipping.Config, logger *zap.SugaredLogger) (Service, error) {
	hc := client.New(fmt.Sprintf("https://%s", config.Shipengine.Host), httpClient, client.WithExternalCall(true))

	return &service{hc, config, logger}, nil
}

type service struct {
	httpClient client.Client
	config     shipping.Config
	logger     *zap.SugaredLogger
}

// GetRatesEstimate returns the estimated rates for the given shipping address and dimensions
// https://shipengine.github.io/shipengine-openapi/#operation/estimate_rates
func (s *service) GetRatesEstimate(ctx context.Context, from, to entities.ShippingAddress, dimensions entities.Dimensions) ([]entities.EstimateRatesResponse, error) {

	// Each Token is associated with different carriers Ids. Use the list GET /v1/carriers endpoint to get the list of carriers
	// Make sure to change the carrierIds on per environment basis
	carriers := strings.Split(s.config.Shipengine.CarrierIds, ",")
	if len(carriers) == 0 {
		return nil, moduleErrors.NewAPIError("SHIPENGINE_MISSING_CARRIERS")
	}

	req := entities.ShippingRateRequest{
		CarrierIds:        carriers,
		FromCountryCode:   from.Country,
		FromPostalCode:    from.Zip,
		FromCityLocality:  from.City,
		FromStateProvince: from.State,
		ToCountryCode:     to.Country,
		ToPostalCode:      to.Zip,
		ToCityLocality:    to.City,
		ToStateProvince:   to.State,
		Weight: entities.Weight{
			Value: dimensions.Weight.InexactFloat64(),
			Unit:  "pound",
		},
		Dimensions: entities.ObjectDimensions{
			Length: dimensions.Length.InexactFloat64(),
			Width:  dimensions.Width.InexactFloat64(),
			Height: dimensions.Height.InexactFloat64(),
			Unit:   "inch",
		},
	}

	var res []entities.EstimateRatesResponse

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

	return res, nil
}

func (s *service) getShipengineApiHeaders() map[string]string {
	return map[string]string{
		"API-Key":      s.config.Shipengine.Token,
		"Content-Type": "application/json",
	}
}

func (s *service) getShipengineReqUrl(path string) string {
	return fmt.Sprintf("/v1/%s", path)
}

package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/nurdsoft/nurd-commerce-core/internal/transport/http/client"
	"github.com/nurdsoft/nurd-commerce-core/shared/cache"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/entities"
	upsConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/ups/config"
	upsEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/ups/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/providers/ups/errors"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

const (
	oauthEndpoint   = "security/v1/oauth/token"
	logPrefix       = "ups"
	sessionCacheKey = "ups-session-cache"
)

var serviceCodeMappings = map[string]struct {
	Code        string
	Description string
}{
	"02": {Code: "ups_second_day_air", Description: "UPS 2nd Day Air"},
	"59": {Code: "ups_second_day_air_am", Description: "UPS 2nd Day Air A.M."},
	"12": {Code: "ups_three_day_select", Description: "UPS 3 Day Select"},
	"03": {Code: "ups_ground", Description: "UPS Ground"},
	"01": {Code: "ups_next_day_air", Description: "UPS Next Day Air"},
	"14": {Code: "ups_next_day_air_early", Description: "UPS Next Day Air Early"},
	"13": {Code: "ups_next_day_air_saver", Description: "UPS Next Day Air Saver"},
	"11": {Code: "ups_standard", Description: "UPS Standard"},
	"07": {Code: "ups_worldwide_express", Description: "UPS Worldwide Express"},
	"08": {Code: "ups_worldwide_expedited", Description: "UPS Worldwide Expedited"},
	"54": {Code: "ups_worldwide_express_plus", Description: "UPS Worldwide Express Plus"},
	"65": {Code: "ups_worldwide_saver", Description: "UPS Worldwide Saver"},
}

type Service interface {
	ValidateAddress(ctx context.Context, address entities.Address) (*entities.Address, error)
	GetShippingRates(ctx context.Context, shipment entities.Shipment) ([]entities.ShippingRate, error)
}

func New(httpClient *http.Client, config upsConfig.Config, log *zap.SugaredLogger, cache cache.Cache) (Service, error) {
	return &service{httpClient, config, log, cache}, nil
}

type service struct {
	httpClient *http.Client
	config     upsConfig.Config
	log        *zap.SugaredLogger
	cache      cache.Cache
}

// GetShippingRates returns the estimated rates for the given shipping address and dimensions
// https://developer.ups.com/tag/Rating?loc=en_US&tag=Rating#operation/Rate
func (s *service) GetShippingRates(ctx context.Context, shipment entities.Shipment) ([]entities.ShippingRate, error) {
	url := fmt.Sprintf("%s/api/rating/v2409/Shop", s.config.APIHost)

	session, err := s.newSession(ctx)
	if err != nil {
		return nil, err
	}

	req := &upsEntities.RateRequestWrapper{
		RateRequest: upsEntities.RateRequest{
			Request: upsEntities.Request{
				TransactionReference: upsEntities.TransactionReference{
					CustomerContext: "Shipping Rate Request",
				},
			},
			Shipment: upsEntities.Shipment{
				Shipper: upsEntities.Party{
					Name:          s.config.ShipperName,
					ShipperNumber: s.config.ShipperNumber,
					Address: upsEntities.Address{
						AddressLine:       []string{shipment.Origin.Address},
						City:              shipment.Origin.City,
						StateProvinceCode: shipment.Origin.StateCode,
						PostalCode:        shipment.Origin.PostalCode,
						CountryCode:       shipment.Origin.CountryCode,
					},
				},
				ShipTo: upsEntities.Party{
					Name: shipment.Destination.FullName,
					Address: upsEntities.Address{
						AddressLine:       []string{shipment.Destination.Address},
						City:              shipment.Destination.City,
						StateProvinceCode: shipment.Destination.StateCode,
						PostalCode:        shipment.Destination.PostalCode,
						CountryCode:       shipment.Destination.CountryCode,
					},
				},
				NumOfPieces: "1", // Assuming only one package will be shipped
				Package: upsEntities.Package{
					PackagingType: upsEntities.CodeDescription{
						Code:        "02",
						Description: "Package",
					},
					Dimensions: upsEntities.Dimensions{
						UnitOfMeasurement: upsEntities.CodeDescription{
							Code:        "IN",
							Description: "Inches",
						},
						Length: shipment.Dimensions.Length.StringFixed(2),
						Width:  shipment.Dimensions.Width.StringFixed(2),
						Height: shipment.Dimensions.Height.StringFixed(2),
					},
					PackageWeight: upsEntities.PackageWeight{
						UnitOfMeasurement: upsEntities.CodeDescription{
							Code:        "LBS",
							Description: "Pounds",
						},
						Weight: shipment.Dimensions.Weight.StringFixed(2),
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(req)
	if err != nil {
		s.log.Error(logPrefix, "json encode failed,", err)
		return nil, err
	}

	data, err := session.httpRequest(http.MethodPost, url, &buf)
	if err != nil {
		s.log.Error(logPrefix, "http request failed,", err)
		return nil, errors.NewAPIError("UPS_RATES_ERROR", err.Error())
	}

	res := &upsEntities.RateResponseWrapper{}
	err = json.Unmarshal(data, res)
	if err != nil {
		s.log.Error(logPrefix, "json decode failed,", err)
		return nil, errors.NewAPIError("UPS_RATES_ERROR", err.Error())
	}

	if res.RateResponse.Response.ResponseStatus.Code != "1" || len(res.RateResponse.RatedShipment) == 0 {
		s.log.Error(logPrefix, "UPS response error:", res.RateResponse.Response.ResponseStatus.Description)
		return nil, errors.NewAPIError("UPS_RATES_ERROR")
	} else {
		var shippingRates []entities.ShippingRate
		for _, rate := range res.RateResponse.RatedShipment {

			totalRate, err := decimal.NewFromString(rate.TotalCharges.MonetaryValue)
			if err != nil {
				s.log.Error(logPrefix, "failed to parse total rate:", err)
				return nil, errors.NewAPIError("UPS_INVALID_RATE")
			}

			mapping := serviceCodeMappings[rate.Service.Code]

			shippingRate := entities.ShippingRate{
				Amount:      totalRate,
				Currency:    rate.TotalCharges.CurrencyCode,
				CarrierName: "UPS",
				CarrierCode: "ups",
				ServiceType: mapping.Description,
				ServiceCode: mapping.Code,
			}
			if rate.GuaranteedDelivery != nil {
				shippingRate.BusinessDaysInTransit = rate.GuaranteedDelivery.BusinessDaysInTransit
			}
			shippingRates = append(shippingRates, shippingRate)
		}

		return shippingRates, nil
	}
}

// ValidateAddress return validation result for the given shipping address
// https://developer.ups.com/tag/Address-Validation?loc=en_US&tag=Rating#operation/AddressValidation
func (s *service) ValidateAddress(ctx context.Context, address entities.Address) (*entities.Address, error) {
	url := fmt.Sprintf("%s/api/addressvalidation/v2/1?regionalrequestindicator=False&maximumcandidatelistsize=10", s.config.APIHost)

	session, err := s.newSession(ctx)
	if err != nil {
		return nil, err
	}

	req := &upsEntities.XAVRequestWrapper{
		XAVRequest: upsEntities.XAVRequest{
			AddressKeyFormat: upsEntities.AddressKeyFormat{
				AddressLine:        []string{address.Address},
				PoliticalDivision1: address.StateCode,
				PoliticalDivision2: address.City,
				PostcodePrimaryLow: address.PostalCode,
				CountryCode:        address.CountryCode,
			},
		},
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(req)
	if err != nil {
		s.log.Error(logPrefix, "json encode failed,", err)
		return nil, err
	}

	data, err := session.httpRequest(http.MethodPost, url, &buf)
	if err != nil {
		s.log.Error(logPrefix, "http request failed,", err)
		return nil, err
	}

	res := &upsEntities.XAVResponseWrapper{}
	err = json.Unmarshal(data, res)
	if err != nil {
		s.log.Error(logPrefix, "json decode failed,", err)
		return nil, err
	}

	if res.XAVResponse.Response.ResponseStatus.Code != "1" || len(res.XAVResponse.Candidate) == 0 {
		s.log.Error(logPrefix, "invalid address response from UPS:", res.XAVResponse.Response.ResponseStatus.Description)
		return nil, errors.NewAPIError("UPS_INVALID_ADDRESS")
	} else {
		candidate := res.XAVResponse.Candidate[0]
		return &entities.Address{
			Address:     candidate.AddressKeyFormat.AddressLine[0],
			City:        candidate.AddressKeyFormat.PoliticalDivision2,
			StateCode:   candidate.AddressKeyFormat.PoliticalDivision1,
			PostalCode:  candidate.AddressKeyFormat.PostcodePrimaryLow,
			CountryCode: candidate.AddressKeyFormat.CountryCode,
		}, nil
	}
}

func (s *service) newSession(ctx context.Context) (*Session, error) {
	var err error
	var res *upsEntities.Oauth2Response

	tokenResCache, err := s.cache.Get(ctx, sessionCacheKey)
	if err == nil {
		err := json.Unmarshal(tokenResCache.([]byte), &res)
		if err != nil {
			s.log.Error("Error unmarshalling cache response")
			return nil, err
		}
	} else {
		res, err = s.getAccessToken(ctx)
		if err != nil {
			s.log.Error(logPrefix, "Failed to get access token from UPS:", err)
			return nil, err
		}
		responseBytes, err := json.Marshal(res)
		if err == nil {
			expiresInSeconds, err := strconv.Atoi(res.ExpiresIn)
			if err != nil {
				s.log.Error("Invalid expires_in value from UPS response")
				return nil, err
			}
			// Subtracting 10 seconds to ensure the token does not expire before it is used
			cacheDuration := time.Duration(expiresInSeconds-10) * time.Second
			_ = s.cache.Set(ctx, sessionCacheKey, responseBytes, cacheDuration)
		}
	}

	sess := &Session{}
	sess.AccessToken = res.AccessToken
	sess.ClientID = res.ClientID
	sess.TokenType = res.TokenType
	sess.IssuedAt = res.IssuedAt
	sess.Status = res.Status
	sess.httpClient = s.httpClient

	return sess, nil
}

// getAccessToken retrieves a new access token from UPS
func (s *service) getAccessToken(ctx context.Context) (*upsEntities.Oauth2Response, error) {
	httpClient := client.New(s.config.SecurityHost, s.httpClient, s.log)

	res := &upsEntities.Oauth2Response{}

	endpoint := oauthEndpoint
	body := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     s.config.ClientID,
		"client_secret": s.config.ClientSecret,
	}

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	basicAuth := "Basic " + basicAuth(s.config.ClientID, s.config.ClientSecret)
	headers["Authorization"] = basicAuth

	err := httpClient.Post(ctx, endpoint, headers, body, res)
	return res, err
}

func basicAuth(clientID, clientSecret string) string {
	auth := clientID + ":" + clientSecret
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

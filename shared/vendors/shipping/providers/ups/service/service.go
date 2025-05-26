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
// https://shipengine.github.io/shipengine-openapi/#operation/estimate_rates
func (s *service) GetShippingRates(ctx context.Context, shipment entities.Shipment) ([]entities.ShippingRate, error) {

	return nil, nil
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

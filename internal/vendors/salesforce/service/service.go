package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nurdsoft/nurd-commerce-core/internal/transport/http/client"
	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/salesforce/config"
	"github.com/nurdsoft/nurd-commerce-core/internal/vendors/salesforce/entities"
	sfErrors "github.com/nurdsoft/nurd-commerce-core/internal/vendors/salesforce/errors"
	"github.com/nurdsoft/nurd-commerce-core/shared/cache"
	"go.uber.org/zap"
)

const (
	oauthEndpoint = "services/oauth2/token"
	logPrefix     = "salesforce"
	cacheKey      = "salesforce-session-cache"
)

type Service interface {
	GetAccountByID(ctx context.Context, accountId string) (*entities.Account, error)
	CreateUserAccount(ctx context.Context, req *entities.CreateSFUserRequest) (*entities.CreateSFUserResponse, error)
	UpdateUserAccount(ctx context.Context, req *entities.UpdateSFUserRequest) error
	CreateUserAddress(ctx context.Context, req *entities.CreateSFAddressRequest) (*entities.CreateSFAddressResponse, error)
	UpdateUserAddress(ctx context.Context, req *entities.UpdateSFAddressRequest) error
	DeleteUserAddress(ctx context.Context, addressId string) error
	CreateProduct(ctx context.Context, req *entities.CreateSFProductRequest) (*entities.CreateSFProductResponse, error)
	CreatePriceBookEntry(ctx context.Context, req *entities.CreateSFPriceBookEntryRequest) (*entities.CreateSFPriceBookEntryResponse, error)
	CreateOrder(ctx context.Context, req *entities.CreateSFOrderRequest) (*entities.CreateSFOrderResponse, error)
	AddOrderItems(ctx context.Context, items []*entities.OrderItem) (*entities.AddOrderItemResponse, error)
	UpdateOrderStatus(ctx context.Context, req *entities.UpdateOrderRequest) error
	GetOrderItems(ctx context.Context, orderId string) (*entities.GetOrderItemsResponse, error)
}
type service struct {
	config     config.Config
	httpClient *http.Client
	log        *zap.SugaredLogger
	cache      cache.Cache
}

func New(cfg config.Config, httpClient *http.Client, logger *zap.SugaredLogger, cache cache.Cache) Service {
	return &service{
		config:     cfg,
		httpClient: httpClient,
		log:        logger,
		cache:      cache,
	}
}

func (s *service) GetAccountByID(ctx context.Context, accountId string) (*entities.Account, error) {
	url := s.makeUrl("Account", accountId)
	session, err := s.newSession(ctx)
	if err != nil {
		s.log.Error("Error creating salesforce session")
		return nil, err
	}

	data, err := session.httpRequest(http.MethodGet, url, nil)
	if err != nil {
		s.log.Error(logPrefix, "http request failed,", err)
		return nil, err
	}

	account := &entities.Account{}

	err = json.Unmarshal(data, account)
	if err != nil {
		s.log.Error(logPrefix, "json decode failed,", err)
		return nil, err
	}

	return account, nil
}

func (s *service) CreateUserAccount(ctx context.Context, account *entities.CreateSFUserRequest) (*entities.CreateSFUserResponse, error) {
	url := s.makeUrl("Account", "")

	session, err := s.newSession(ctx)
	if err != nil {
		s.log.Error("Error creating salesforce session")
		return nil, err
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(account)
	if err != nil {
		s.log.Error(logPrefix, "json encode failed,", err)
		return nil, err
	}

	data, err := session.httpRequest(http.MethodPost, url, &buf)
	if err != nil {
		s.log.Error(logPrefix, "http request failed,", err)
		return nil, err
	}

	res := &entities.CreateSFUserResponse{}
	err = json.Unmarshal(data, res)
	if err != nil {
		s.log.Error(logPrefix, "json decode failed,", err)
		return nil, err
	}

	return res, nil
}

func (s *service) UpdateUserAccount(ctx context.Context, account *entities.UpdateSFUserRequest) error {
	url := s.makeUrl("Account", account.ID)
	session, err := s.newSession(ctx)
	if err != nil {
		s.log.Error("Error creating salesforce session")
		return err
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(account)
	if err != nil {
		s.log.Error(logPrefix, "json encode failed,", err)
		return err
	}

	_, err = session.httpRequest(http.MethodPatch, url, &buf)
	if err != nil {
		s.log.Error(logPrefix, "http request failed,", err)
		return err
	}
	return nil
}

func (s *service) CreateUserAddress(ctx context.Context, address *entities.CreateSFAddressRequest) (*entities.CreateSFAddressResponse, error) {
	url := s.makeUrl("Address__c", "")

	session, err := s.newSession(ctx)
	if err != nil {
		s.log.Error("Error creating salesforce session")
		return nil, err
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(address)
	if err != nil {
		s.log.Error(logPrefix, "json encode failed,", err)
		return nil, err
	}

	data, err := session.httpRequest(http.MethodPost, url, &buf)
	if err != nil {
		s.log.Error(logPrefix, "http request failed,", err)
		return nil, err
	}

	res := &entities.CreateSFAddressResponse{}
	err = json.Unmarshal(data, res)
	if err != nil {
		s.log.Error(logPrefix, "json decode failed,", err)
		return nil, err
	}

	return res, nil
}

func (s *service) DeleteUserAddress(ctx context.Context, addressId string) error {
	url := s.makeUrl("Address__c", addressId)
	session, err := s.newSession(ctx)
	if err != nil {
		s.log.Error("Error creating salesforce session")
		return err
	}

	_, err = session.httpRequest(http.MethodDelete, url, nil)
	if err != nil {
		s.log.Error(logPrefix, "http request failed,", err)
		return err
	}

	return nil
}

func (s *service) UpdateUserAddress(ctx context.Context, address *entities.UpdateSFAddressRequest) error {
	url := s.makeUrl("Address__c", address.AddressID)
	session, err := s.newSession(ctx)
	if err != nil {
		s.log.Error("Error creating salesforce session")
		return err
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(address)
	if err != nil {
		s.log.Error(logPrefix, "json encode failed,", err)
		return err
	}

	_, err = session.httpRequest(http.MethodPatch, url, &buf)
	if err != nil {
		s.log.Error(logPrefix, "http request failed,", err)
		return err
	}

	return nil
}

func (s *service) CreateProduct(ctx context.Context, product *entities.CreateSFProductRequest) (*entities.CreateSFProductResponse, error) {
	url := s.makeUrl("Product2", "")

	session, err := s.newSession(ctx)
	if err != nil {
		s.log.Error("Error creating salesforce session")
		return nil, err
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(product)
	if err != nil {
		s.log.Error(logPrefix, "json encode failed,", err)
		return nil, err
	}

	data, err := session.httpRequest(http.MethodPost, url, &buf)
	if err != nil {
		s.log.Error(logPrefix, "http request failed,", err)
		return nil, err
	}

	res := &entities.CreateSFProductResponse{}
	err = json.Unmarshal(data, res)
	if err != nil {
		s.log.Error(logPrefix, "json decode failed,", err)
		return nil, err
	}

	return res, nil
}

func (s *service) CreatePriceBookEntry(ctx context.Context, pricebook *entities.CreateSFPriceBookEntryRequest) (*entities.CreateSFPriceBookEntryResponse, error) {
	url := s.makeUrl("PricebookEntry", "")

	session, err := s.newSession(ctx)
	if err != nil {
		s.log.Error("Error creating salesforce session")
		return nil, err
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(pricebook)
	if err != nil {
		s.log.Error(logPrefix, "json encode failed,", err)
		return nil, err
	}

	data, err := session.httpRequest(http.MethodPost, url, &buf)
	if err != nil {
		s.log.Error(logPrefix, "http request failed,", err)
		return nil, err
	}

	res := &entities.CreateSFPriceBookEntryResponse{}
	err = json.Unmarshal(data, res)
	if err != nil {
		s.log.Error(logPrefix, "json decode failed,", err)
		return nil, err
	}

	return res, nil
}

func (s *service) CreateOrder(ctx context.Context, order *entities.CreateSFOrderRequest) (*entities.CreateSFOrderResponse, error) {
	url := s.makeUrl("Order", "")

	session, err := s.newSession(ctx)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(order)
	if err != nil {
		s.log.Error(logPrefix, "json encode failed,", err)
		return nil, err
	}

	data, err := session.httpRequest(http.MethodPost, url, &buf)
	if err != nil {
		s.log.Error(logPrefix, "http request failed,", err)
		return nil, err
	}

	res := &entities.CreateSFOrderResponse{}
	err = json.Unmarshal(data, res)
	if err != nil {
		s.log.Error(logPrefix, "json decode failed,", err)
		return nil, err
	}

	return res, nil
}

func (s *service) AddOrderItems(ctx context.Context, items []*entities.OrderItem) (*entities.AddOrderItemResponse, error) {
	url := fmt.Sprintf("%s/services/data/%s/composite/batch", s.config.ApiHost, s.config.ApiVersion)

	batchRequest := &entities.AddOrderItemRequest{}

	for _, item := range items {
		batchRequest.BatchRequests = append(batchRequest.BatchRequests, entities.BatchRequests{
			Method:    "POST",
			URL:       fmt.Sprintf("/services/data/%s/sobjects/OrderItem", s.config.ApiVersion),
			RichInput: *item,
		})
	}

	session, err := s.newSession(ctx)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(batchRequest)
	if err != nil {
		s.log.Error(logPrefix, "json encode failed,", err)
		return nil, err
	}

	data, err := session.httpRequest(http.MethodPost, url, &buf)
	if err != nil {
		s.log.Error(logPrefix, "http request failed,", err)
		return nil, err
	}

	res := &entities.AddOrderItemResponse{}
	err = json.Unmarshal(data, res)
	if err != nil {
		s.log.Error(logPrefix, "json decode failed,", err)
		return nil, err
	}
	return res, nil
}

func (s *service) UpdateOrderStatus(ctx context.Context, order *entities.UpdateOrderRequest) error {
	url := s.makeUrl("Order", order.OrderId)
	session, err := s.newSession(ctx)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(order)
	if err != nil {
		s.log.Error(logPrefix, "json encode failed,", err)
		return err
	}

	_, err = session.httpRequest(http.MethodPatch, url, &buf)
	if err != nil {
		s.log.Error(logPrefix, "http request failed,", err)
		return err
	}

	return nil
}

func (s *service) newSession(ctx context.Context) (*SFSession, error) {
	var err error
	var res *entities.Oauth2Response

	tokenResCache, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		err := json.Unmarshal(tokenResCache.([]byte), &res)
		if err != nil {
			s.log.Error("Error unmarshalling cache response")
			return nil, err
		}
	} else {
		res, err = s.getAccessToken(ctx)
		if err != nil {
			return nil, err
		}
		responseBytes, err := json.Marshal(res)
		if err == nil {
			// TODO: check how long the token is valid for and set the cache duration accordingly
			_ = s.cache.Set(ctx, cacheKey, responseBytes, 15*time.Minute)
		}
	}

	sess := &SFSession{}
	sess.AccessToken = res.AccessToken
	sess.InstanceURL = res.InstanceURL
	sess.ID = res.ID
	sess.TokenType = res.TokenType
	sess.IssuedAt = res.IssuedAt
	sess.Signature = res.Signature
	sess.httpClient = s.httpClient

	return sess, nil
}

// getAccessToken retrieves a new access token from salesforce
func (s *service) getAccessToken(ctx context.Context) (*entities.Oauth2Response, error) {
	httpClient := client.New(s.config.ApiHost, s.httpClient)

	res := &entities.Oauth2Response{}

	url := fmt.Sprintf("%s?grant_type=password&client_id=%s&client_secret=%s&username=%s&password=%s",
		oauthEndpoint, s.config.ClientID, s.config.ClientSecret, s.config.Username, s.config.Password)

	err := httpClient.Post(ctx, url, nil, nil, res)

	if res.Error != "" {
		return nil, &sfErrors.ErrSalesforceError{
			Message:      res.Error,
			HttpCode:     http.StatusInternalServerError,
			ErrorCode:    "",
			ErrorMessage: fmt.Sprintf("salesforce error:%s", res.ErrorDescription),
		}
	}
	return res, err
}

func (s *service) GetOrderItems(ctx context.Context, orderId string) (*entities.GetOrderItemsResponse, error) {
	url := fmt.Sprintf(
		"%s/services/data/%s/query/?q=SELECT+Id,Quantity,Product2Id,Description,PricebookEntryId,Type__c+FROM+OrderItem+WHERE+OrderId='%s'",
		s.config.ApiHost,
		s.config.ApiVersion,
		orderId,
	)

	session, err := s.newSession(ctx)
	if err != nil {
		return nil, err
	}

	data, err := session.httpRequest(http.MethodGet, url, nil)
	if err != nil {
		s.log.Error(logPrefix, "http request failed,", err)
		return nil, err
	}

	res := &entities.GetOrderItemsResponse{}
	err = json.Unmarshal(data, res)
	if err != nil {
		s.log.Error(logPrefix, "json decode failed,", err)
		return nil, err
	}

	return res, nil
}

// makeUrl creates a salesforce api url with the given sObjectType and objectId
func (s *service) makeUrl(sObjectType, objectId string) string {
	url := fmt.Sprintf("%s/services/data/%s/sobjects/%s/%s",
		s.config.ApiHost, s.config.ApiVersion, sObjectType, objectId)

	return url
}

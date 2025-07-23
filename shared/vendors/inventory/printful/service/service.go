package service

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/nurdsoft/nurd-commerce-core/internal/transport/http/client"
	printfulConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/printful/config"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/printful/entities"
	"go.uber.org/zap"
)

const (
	logPrefix = "printful"
)

type Service interface {
	GetSyncProducts(ctx context.Context, req entities.GetSyncProductsRequest) (*entities.SyncProductsResponse, error)
	GetSyncProduct(ctx context.Context, productID int) (*entities.GetSyncProductResponse, error)
}

type service struct {
	config     printfulConfig.Config
	httpClient *http.Client
	log        *zap.SugaredLogger
}

func New(config printfulConfig.Config, httpClient *http.Client, log *zap.SugaredLogger) Service {
	return &service{
		config:     config,
		httpClient: httpClient,
		log:        log,
	}
}

func (s *service) GetSyncProducts(ctx context.Context, req entities.GetSyncProductsRequest) (*entities.SyncProductsResponse, error) {
	httpClient := client.New(s.config.BaseURL, s.httpClient, s.log)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.config.OAuthToken),
	}

	queryParams := url.Values{}
	if req.Search != "" {
		queryParams.Add("search", req.Search)
	}
	if req.Limit > 0 {
		queryParams.Add("limit", strconv.Itoa(req.Limit))
	}
	if req.Offset > 0 {
		queryParams.Add("offset", strconv.Itoa(req.Offset))
	}

	url := "sync/products"
	if len(queryParams) > 0 {
		url = fmt.Sprintf("sync/products?%s", queryParams.Encode())
	}

	response := &entities.SyncProductsResponse{}

	err := httpClient.Get(ctx, url, headers, nil, response)
	if err != nil {
		s.log.Error(logPrefix, "Failed to get sync products:", err)
		return nil, err
	}

	return response, nil
}

func (s *service) GetSyncProduct(ctx context.Context, productID int) (*entities.GetSyncProductResponse, error) {
	httpClient := client.New(s.config.BaseURL, s.httpClient, s.log)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.config.OAuthToken),
	}

	url := fmt.Sprintf("sync/products/%d", productID)

	response := &entities.GetSyncProductResponse{}

	err := httpClient.Get(ctx, url, headers, nil, response)
	if err != nil {
		s.log.Error(logPrefix, "Failed to get sync product:", err)
		return nil, err
	}

	return response, nil
}
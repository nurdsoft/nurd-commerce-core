package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/config"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/entities"
)

const (
	AuthorizeNetStatusApproved      = "approved"
	AuthorizeNetStatusDeclined      = "declined"
	AuthorizeNetStatusError         = "error"
	AuthorizeNetStatusHeldForReview = "held_for_review"
	AuthorizeNetStatusUnknown       = "unknown"
)

type Service interface {
	CreateCustomerProfile(ctx context.Context, req entities.CreateCustomerRequest) (entities.CreateCustomerResponse, error)
	CreateCustomerPaymentProfile(ctx context.Context, req entities.CreateCustomerPaymentProfileRequest) (entities.CreateCustomerPaymentProfileResponse, error)
	GetCustomerPaymentProfiles(ctx context.Context, req entities.GetPaymentProfilesRequest) (entities.GetPaymentProfilesResponse, error)
	CreatePaymentTransaction(ctx context.Context, req entities.CreatePaymentTransactionRequest) (entities.CreatePaymentTransactionResponse, error)
}

func New(cfg config.Config, logger *zap.SugaredLogger) Service {
	transport := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	validationMode := "testMode"
	if cfg.LiveMode {
		validationMode = "liveMode"
	}

	return &service{
		apiLoginID:     cfg.ApiLoginID,
		transactionKey: cfg.TransactionKey,
		endpoint:       cfg.Endpoint,
		validationMode: validationMode,
		httpClient:     httpClient,
		logger:         logger,
	}
}

type service struct {
	apiLoginID     string
	transactionKey string
	endpoint       string
	validationMode string
	httpClient     *http.Client
	logger         *zap.SugaredLogger
}

func (s *service) CreateCustomerProfile(ctx context.Context, req entities.CreateCustomerRequest) (entities.CreateCustomerResponse, error) {
	s.logger.Infof("Creating customer profile: customerID=%s", req.CustomerID)
	requestData := CreateCustomerProfileRequest{
		Data: CreateCustomerProfileRequestData{
			MerchantAuthentication: merchantAuthentication{
				Name:           s.apiLoginID,
				TransactionKey: s.transactionKey,
			},
			Profile: Profile{
				MerchantCustomerID: req.CustomerID[:20], // Authorize.net has a 20 chars limit for customer ID
				Description:        req.Description,
				Email:              req.Email,
			},
		},
	}

	var response CreateCustomerProfileResponse
	if err := s.sendRequest(ctx, requestData, &response); err != nil {
		s.logger.Errorf("Failed to create customer profile (sendRequest): %v", err)
		return entities.CreateCustomerResponse{}, fmt.Errorf("failed to create customer profile: %w", err)
	}

	if err := checkResponseForErrors(response.Messages); err != nil {
		s.logger.Errorf("authorize.net API error when creating customer profile: %v", err)
		return entities.CreateCustomerResponse{}, err
	}

	return entities.CreateCustomerResponse{ProfileID: response.CustomerProfileID}, nil
}

func (s *service) CreateCustomerPaymentProfile(ctx context.Context, req entities.CreateCustomerPaymentProfileRequest) (entities.CreateCustomerPaymentProfileResponse, error) {
	s.logger.Infof("Creating customer payment profile: profileID=%s", req.ProfileID)
	requestData := CreateCustomerPaymentProfileRequest{
		Data: CreateCustomerPaymentProfileRequestData{
			MerchantAuthentication: merchantAuthentication{
				Name:           s.apiLoginID,
				TransactionKey: s.transactionKey,
			},
			CustomerProfileID: req.ProfileID,
			PaymentProfile: PaymentProfile{
				Payment: Payment{
					CreditCard: CreditCard{
						CardNumber:     req.CardNumber,
						ExpirationDate: req.ExpirationDate,
					},
				},
				DefaultPaymentProfile: true,
			},
			ValidationMode: s.validationMode,
		},
	}

	var response CreateCustomerPaymentProfileResponse
	if err := s.sendRequest(ctx, requestData, &response); err != nil {
		s.logger.Errorf("Failed to create customer payment profile (sendRequest): %v", err)
		return entities.CreateCustomerPaymentProfileResponse{}, fmt.Errorf("failed to create customer profile: %w", err)
	}

	if err := checkResponseForErrors(response.Messages); err != nil {
		s.logger.Errorf("authorize.net API error when creating customer payment profile: %v", err)
		return entities.CreateCustomerPaymentProfileResponse{}, err
	}

	return entities.CreateCustomerPaymentProfileResponse{
		ProfileID:        response.CustomerProfileID,
		PaymentProfileID: response.CustomerPaymentProfileID,
	}, nil
}

func (s *service) GetCustomerPaymentProfiles(ctx context.Context, req entities.GetPaymentProfilesRequest) (entities.GetPaymentProfilesResponse, error) {
	s.logger.Infof("Getting customer payment profiles: profileID=%s", req.ProfileID)
	requestData := GetCustomerProfileRequest{
		Data: GetCustomerProfileRequestData{
			MerchantAuthentication: merchantAuthentication{
				Name:           s.apiLoginID,
				TransactionKey: s.transactionKey,
			},
			CustomerProfileIID:   req.ProfileID,
			UnmaskExpirationDate: true,
		},
	}

	var response GetCustomerProfileResponse
	if err := s.sendRequest(ctx, requestData, &response); err != nil {
		s.logger.Errorf("Failed to get customer payment profiles (sendRequest): %v", err)
		return entities.GetPaymentProfilesResponse{}, fmt.Errorf("failed to get customer payment profiles: %w", err)
	}

	if err := checkResponseForErrors(response.Messages); err != nil {
		s.logger.Errorf("authorize.net API error when getting customer payment profiles: %v", err)
		return entities.GetPaymentProfilesResponse{}, err
	}

	if len(response.Profile.PaymentProfiles) == 0 {
		s.logger.Info("No payment profiles found for customer profile")
		return entities.GetPaymentProfilesResponse{
			PaymentProfiles: []entities.PaymentProfile{},
		}, nil
	}

	paymentProfiles := make([]entities.PaymentProfile, 0, len(response.Profile.PaymentProfiles))
	for _, profile := range response.Profile.PaymentProfiles {
		paymentProfiles = append(paymentProfiles, entities.PaymentProfile{
			ID:             profile.CustomerPaymentProfileID,
			CardNumber:     profile.Payment.CreditCard.CardNumber,
			CardType:       profile.Payment.CreditCard.CardType,
			ExpirationDate: profile.Payment.CreditCard.ExpirationDate,
		})
	}

	return entities.GetPaymentProfilesResponse{
		PaymentProfiles: paymentProfiles,
	}, nil
}

func (s *service) CreatePaymentTransaction(ctx context.Context, req entities.CreatePaymentTransactionRequest) (entities.CreatePaymentTransactionResponse, error) {
	amount := req.Amount.StringFixed(2)
	s.logger.Infof("Creating payment transaction: profileID=%s", req.ProfileID)

	requestData := CreateTransactionRequest{
		Data: TransactionRequestData{
			MerchantAuthentication: merchantAuthentication{
				Name:           s.apiLoginID,
				TransactionKey: s.transactionKey,
			},
			TransactionRequest: TransactionRequest{
				TransactionType: "authCaptureTransaction",
				Amount:          amount,
				Payment: PaymentNonce{
					OpaqueData: OpaqueData{
						DataDescriptor: "COMMON.ACCEPT.INAPP.PAYMENT",
						DataValue:      req.PaymentNonce,
					},
				},
				Customer: Customer{
					ID: req.ProfileID,
				},
			},
		},
	}

	var response CreateTransactionResponse
	if err := s.sendRequest(ctx, requestData, &response); err != nil {
		s.logger.Errorf("Failed to create transaction (sendRequest): %v", err)
		return entities.CreatePaymentTransactionResponse{}, fmt.Errorf("failed to create transaction: %w", err)
	}

	if err := checkResponseForErrors(response.Messages); err != nil {
		s.logger.Errorf("authorize.net API error when creating transaction: %v", err)
		return entities.CreatePaymentTransactionResponse{}, err
	}

	status := mapResponseCodeToStatus(response.TransactionResponse.ResponseCode)
	return entities.CreatePaymentTransactionResponse{
		ID:     response.TransactionResponse.TransID,
		Status: status,
	}, nil
}

func mapResponseCodeToStatus(responseCode string) string {
	switch responseCode {
	case "1":
		return AuthorizeNetStatusApproved
	case "2":
		return AuthorizeNetStatusDeclined
	case "3":
		return AuthorizeNetStatusError
	case "4":
		return AuthorizeNetStatusHeldForReview
	default:
		return AuthorizeNetStatusUnknown
	}
}

func checkResponseForErrors(messages Messages) error {
	if messages.ResultCode != "Ok" {
		if len(messages.Message) == 0 {
			return fmt.Errorf("unknown error occurred")
		}
		return fmt.Errorf("authorize.net API error: %s - %s", messages.Message[0].Code, messages.Message[0].Text)
	}
	return nil
}

func (s *service) sendRequest(ctx context.Context, requestData any, responseData any) error {
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// Remove BOM if present: responses from Authorize.net start with a BOM (byte order mark)
	body = removeBOM(body)

	err = json.Unmarshal(body, responseData)
	if err != nil {
		return fmt.Errorf("error parsing response: %w", err)
	}

	return nil
}

// Response body has a ZWNBSP char prefix: remove it
func removeBOM(data []byte) []byte {
	s := string(data)

	if strings.HasPrefix(s, "\uFEFF") {
		return []byte(strings.TrimPrefix(s, "\uFEFF"))
	}

	return data
}

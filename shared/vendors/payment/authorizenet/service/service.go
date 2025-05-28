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

	endpoint := "https://apitest.authorize.net/xml/v1/request.api"
	validationMode := "testMode"
	if cfg.LiveMode {
		endpoint = "https://api.authorize.net/xml/v1/request.api"
		validationMode = "liveMode"
	}

	return &service{
		apiLoginID:     cfg.ApiLoginID,
		transactionKey: cfg.TransactionKey,
		endpoint:       endpoint,
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
	s.logger.Infof("Creating customer profile: customerID=%s, email=%s", req.CustomerID, req.Email)
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

	response, err := s.sendRequest(ctx, requestData)
	if err != nil {
		s.logger.Errorf("Failed to create customer profile (sendRequest): %v", err)
		return entities.CreateCustomerResponse{}, fmt.Errorf("failed to create customer profile: %w", err)
	}

	if err = checkResponseForErrors(response); err != nil {
		s.logger.Errorf("API error when creating customer profile: %v", err)
		return entities.CreateCustomerResponse{}, err
	}

	customerProfileID, ok := response["customerProfileId"].(string)
	if !ok {
		s.logger.Error("Customer profile ID not found in response")
		return entities.CreateCustomerResponse{}, fmt.Errorf("customer profile ID not found in response")
	}

	return entities.CreateCustomerResponse{ProfileID: customerProfileID}, nil
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
			ValidationMode: s.validationMode, // Options: "none", "testMode", "liveMode"
		},
	}

	response, err := s.sendRequest(ctx, requestData)
	if err != nil {
		s.logger.Errorf("Failed to create customer payment profile (sendRequest): %v", err)
		return entities.CreateCustomerPaymentProfileResponse{}, fmt.Errorf("failed to create customer profile: %w", err)
	}

	if err = checkResponseForErrors(response); err != nil {
		s.logger.Errorf("API error when creating customer payment profile: %v", err)
		return entities.CreateCustomerPaymentProfileResponse{}, err
	}

	paymentProfileID, ok := response["customerPaymentProfileId"].(string)
	if !ok {
		s.logger.Error("Customer payment profile ID not found in response")
		return entities.CreateCustomerPaymentProfileResponse{}, fmt.Errorf("customer payment profile ID not found in response")
	}
	customerProfileID, _ := response["customerProfileId"].(string)

	return entities.CreateCustomerPaymentProfileResponse{
		ProfileID:        customerProfileID,
		PaymentProfileID: paymentProfileID,
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

	response, err := s.sendRequest(ctx, requestData)
	if err != nil {
		s.logger.Errorf("Failed to get customer payment profiles (sendRequest): %v", err)
		return entities.GetPaymentProfilesResponse{}, fmt.Errorf("failed to get customer payment profiles: %w", err)
	}

	if err = checkResponseForErrors(response); err != nil {
		s.logger.Errorf("API error when getting customer payment profiles: %v", err)
		return entities.GetPaymentProfilesResponse{}, err
	}

	profileResponse, ok := response["profile"].(map[string]any)
	if !ok {
		s.logger.Error("Profile not found in response (get payment methods)")
		return entities.GetPaymentProfilesResponse{}, fmt.Errorf("profile not found in response")
	}

	paymentProfilesResp, ok := profileResponse["paymentProfiles"].([]any)
	if !ok {
		s.logger.Info("No payment profiles found for customer profile")
		return entities.GetPaymentProfilesResponse{
			PaymentProfiles: []entities.PaymentProfile{},
		}, nil
	}

	paymentProfiles := make([]entities.PaymentProfile, 0, len(paymentProfilesResp))
	for _, paymentProfile := range paymentProfilesResp {
		profileMap, ok := paymentProfile.(map[string]any)
		if !ok {
			continue
		}

		paymentMap, ok := profileMap["payment"].(map[string]any)
		if !ok {
			continue
		}

		creditCardMap, ok := paymentMap["creditCard"].(map[string]any)
		if !ok {
			continue
		}

		paymentProfiles = append(paymentProfiles, entities.PaymentProfile{
			ID:             profileMap["customerPaymentProfileId"].(string),
			CardNumber:     creditCardMap["cardNumber"].(string),
			CardType:       creditCardMap["cardType"].(string),
			ExpirationDate: creditCardMap["expirationDate"].(string),
		})
	}

	return entities.GetPaymentProfilesResponse{
		PaymentProfiles: paymentProfiles,
	}, nil
}

func (s *service) CreatePaymentTransaction(ctx context.Context, req entities.CreatePaymentTransactionRequest) (entities.CreatePaymentTransactionResponse, error) {
	amount := req.Amount.StringFixed(2)
	s.logger.Infof("Creating payment transaction: profileID=%s, amount=%s", req.ProfileID, amount)

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
				// BillTo: BillTo{
				// 	Zip: "46282",
				// },
			},
		},
	}

	response, err := s.sendRequest(ctx, requestData)
	if err != nil {
		s.logger.Errorf("Failed to create transaction (sendRequest): %v", err)
		return entities.CreatePaymentTransactionResponse{}, fmt.Errorf("failed to create transaction: %w", err)
	}

	transID, status, err := extractTransactionDetails(response)
	if err != nil {
		s.logger.Errorf("Failed to extract transaction details: %v", err)
		return entities.CreatePaymentTransactionResponse{}, err
	}

	return entities.CreatePaymentTransactionResponse{
		ID:     transID,
		Status: status,
	}, nil
}

func extractTransactionDetails(response map[string]any) (string, string, error) {
	transactionResponse, ok := response["transactionResponse"].(map[string]any)
	if !ok {
		return "", "", fmt.Errorf("transaction response not found")
	}

	transID, ok := transactionResponse["transId"].(string)
	if !ok {
		return "", "", fmt.Errorf("transaction ID not found")
	}

	responseCode, ok := transactionResponse["responseCode"].(string)
	if !ok {
		return "", "", fmt.Errorf("response code not found")
	}

	var status string
	switch responseCode {
	case "1":
		status = AuthorizeNetStatusApproved
	case "2":
		status = AuthorizeNetStatusDeclined
	case "3":
		status = AuthorizeNetStatusError
	case "4":
		status = AuthorizeNetStatusHeldForReview
	default:
		status = AuthorizeNetStatusUnknown
	}

	return transID, status, nil
}

func checkResponseForErrors(response map[string]any) error {
	messagesObj, ok := response["messages"].(map[string]any)
	if !ok {
		return fmt.Errorf("messages object not found in response")
	}

	resultCode, ok := messagesObj["resultCode"].(string)
	if !ok {
		return fmt.Errorf("result code not found in response")
	}

	if resultCode != "Ok" {
		messageList, ok := messagesObj["message"].([]any)
		if !ok || len(messageList) == 0 {
			return fmt.Errorf("unknown error occurred")
		}

		message, ok := messageList[0].(map[string]any)
		if !ok {
			return fmt.Errorf("unknown error occurred")
		}

		code, _ := message["code"].(string)
		text, _ := message["text"].(string)

		return fmt.Errorf("API error: %s - %s", code, text)
	}

	return nil
}

func (s *service) sendRequest(ctx context.Context, requestData any) (map[string]any, error) {
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	fmt.Println(string(body))

	// Remove BOM if present: responses from Authorize.net start with a BOM (byte order mark)
	body = removeBOM(body)

	var responseData map[string]any
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return responseData, nil
}

// Response body has a ZWNBSP char prefix: remove it
func removeBOM(data []byte) []byte {
	s := string(data)

	if strings.HasPrefix(s, "\uFEFF") {
		return []byte(strings.TrimPrefix(s, "\uFEFF"))
	}

	return data
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/entities"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCreateCustomerProfile(t *testing.T) {
	ctx := context.TODO()
	customerID := uuid.NewString()
	apiLoginID := "test-login"
	transactionKey := "test-key"

	t.Run("Success", func(t *testing.T) {
		expectedProfileID := "523516027"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method and content type
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var requestBody CreateCustomerProfileRequest
			err := json.NewDecoder(r.Body).Decode(&requestBody)
			assert.NoError(t, err)

			// Verify request structure
			assert.Equal(t, customerID, requestBody.Data.Profile.MerchantCustomerID)
			assert.Equal(t, "Test Customer", requestBody.Data.Profile.Description)
			assert.Equal(t, "test@example.com", requestBody.Data.Profile.Email)
			assert.Equal(t, apiLoginID, requestBody.Data.MerchantAuthentication.Name)
			assert.Equal(t, transactionKey, requestBody.Data.MerchantAuthentication.TransactionKey)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(fmt.Appendf(nil, `{
				"customerProfileId": "%s",
				"customerPaymentProfileIdList": [],
				"customerShippingAddressIdList": [],
				"validationDirectResponseList": [],
				"messages": {
					"resultCode": "Ok",
					"message": [
						{
							"code": "I00001",
							"text": "Successful."
						}
					]
				}
			}`, expectedProfileID))
		}))
		defer server.Close()

		svc := &service{
			apiLoginID:     apiLoginID,
			transactionKey: transactionKey,
			endpoint:       server.URL,
			httpClient:     &http.Client{},
			logger:         zap.NewExample().Sugar(),
		}

		req := entities.CreateCustomerRequest{
			CustomerID:  customerID,
			Description: "Test Customer",
			Email:       "test@example.com",
		}

		res, err := svc.CreateCustomerProfile(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, expectedProfileID, res.ProfileID)
	})

	t.Run("API Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK) // API returns 200 even for errors
			w.Write([]byte(`{
				"messages": {
					"resultCode": "Error",
					"message": [
						{
							"code": "E00003",
							"text": "The 'AnetApi/xml/v1/schema/AnetApiSchema.xsd:merchantCustomerId' element is invalid - The value &#39;402e3478-9977-42fd-ac&#39; is invalid according to its datatype 'String' - The actual length is greater than the MaxLength value."
						}
					]
				}
			}`))
		}))
		defer server.Close()

		svc := &service{
			apiLoginID:     apiLoginID,
			transactionKey: transactionKey,
			endpoint:       server.URL,
			httpClient:     &http.Client{},
			logger:         zap.NewExample().Sugar(),
		}

		req := entities.CreateCustomerRequest{
			CustomerID:  customerID,
			Description: "Test Customer",
			Email:       "test@example.com",
		}

		res, err := svc.CreateCustomerProfile(ctx, req)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "E00003")
		assert.Empty(t, res)
	})

	t.Run("HTTP Status Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		svc := &service{
			apiLoginID:     apiLoginID,
			transactionKey: transactionKey,
			endpoint:       server.URL,
			httpClient:     &http.Client{},
			logger:         zap.NewExample().Sugar(),
		}

		req := entities.CreateCustomerRequest{
			CustomerID:  customerID,
			Description: "Test Customer",
			Email:       "test@example.com",
		}

		res, err := svc.CreateCustomerProfile(ctx, req)

		assert.Error(t, err)
		assert.Empty(t, res)
	})
}

func TestCreateCustomerPaymentProfile(t *testing.T) {
	ctx := context.TODO()
	profileID := "test-profile-id"
	apiLoginID := "test-login"
	transactionKey := "test-key"
	cardNumber := "4111111111111111"
	expirationDate := "2025-12"
	expectedPaymentID := "pm_1"

	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var requestBody CreateCustomerPaymentProfileRequest
			err := json.NewDecoder(r.Body).Decode(&requestBody)
			assert.NoError(t, err)

			assert.Equal(t, apiLoginID, requestBody.Data.MerchantAuthentication.Name)
			assert.Equal(t, transactionKey, requestBody.Data.MerchantAuthentication.TransactionKey)
			assert.Equal(t, profileID, requestBody.Data.CustomerProfileID)
			assert.Equal(t, cardNumber, requestBody.Data.PaymentProfile.Payment.CreditCard.CardNumber)
			assert.Equal(t, expirationDate, requestBody.Data.PaymentProfile.Payment.CreditCard.ExpirationDate)
			assert.Equal(t, "testMode", requestBody.Data.ValidationMode)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			w.Write(fmt.Appendf(nil, `{
				"customerProfileId": "%s",
				"customerPaymentProfileId": "%s",
				"messages": {
					"resultCode": "Ok",
					"message": [
						{
							"code": "I00001",
							"text": "Successful."
						}
					]
				}
			}`, profileID, expectedPaymentID))
		}))
		defer server.Close()

		svc := &service{
			apiLoginID:     apiLoginID,
			transactionKey: transactionKey,
			endpoint:       server.URL,
			httpClient:     &http.Client{},
			logger:         zap.NewExample().Sugar(),
			validationMode: "testMode",
		}

		req := entities.CreateCustomerPaymentProfileRequest{
			ProfileID:      profileID,
			CardNumber:     cardNumber,
			ExpirationDate: expirationDate,
		}

		res, err := svc.CreateCustomerPaymentProfile(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, expectedPaymentID, res.PaymentProfileID)
		assert.Equal(t, profileID, res.ProfileID)
	})

	t.Run("API Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"messages": {
					"resultCode": "Error",
					"message": [
						{
							"code": "E00013",
							"text": "Validation Mode is invalid without Payment Profiles."
						}
					]
				}
			}`))
		}))
		defer server.Close()

		svc := &service{
			apiLoginID:     apiLoginID,
			transactionKey: transactionKey,
			endpoint:       server.URL,
			httpClient:     &http.Client{},
			logger:         zap.NewExample().Sugar(),
			validationMode: "testMode",
		}

		req := entities.CreateCustomerPaymentProfileRequest{
			ProfileID:      profileID,
			CardNumber:     cardNumber,
			ExpirationDate: expirationDate,
		}

		res, err := svc.CreateCustomerPaymentProfile(ctx, req)

		assert.Error(t, err)
		assert.Empty(t, res)
		assert.ErrorContains(t, err, "E00013")
	})

	t.Run("HTTP Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		svc := &service{
			apiLoginID:     apiLoginID,
			transactionKey: transactionKey,
			endpoint:       server.URL,
			httpClient:     &http.Client{},
			logger:         zap.NewExample().Sugar(),
		}

		req := entities.CreateCustomerPaymentProfileRequest{
			ProfileID:      profileID,
			CardNumber:     cardNumber,
			ExpirationDate: expirationDate,
		}

		res, err := svc.CreateCustomerPaymentProfile(ctx, req)

		assert.Error(t, err)
		assert.Empty(t, res)
	})
}

func TestGetCustomerPaymentProfiles(t *testing.T) {
	ctx := context.TODO()
	profileID := "test-profile-id"
	apiLoginID := "test-login"
	transactionKey := "test-key"

	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var requestBody GetCustomerProfileRequest
			err := json.NewDecoder(r.Body).Decode(&requestBody)
			assert.NoError(t, err)
			assert.Equal(t, profileID, requestBody.Data.CustomerProfileIID)
			assert.Equal(t, apiLoginID, requestBody.Data.MerchantAuthentication.Name)
			assert.Equal(t, transactionKey, requestBody.Data.MerchantAuthentication.TransactionKey)
			assert.Equal(t, true, requestBody.Data.UnmaskExpirationDate)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"profile": {
					"paymentProfiles": [{
						"defaultPaymentProfile": true,
						"customerPaymentProfileId": "535672235",
						"payment": {
							"creditCard": {
								"cardNumber": "XXXX8888",
								"expirationDate": "2025-12", 
								"cardType": "Visa"
							}
						}
					}],
					"profileType": "regular",
					"customerProfileId": "523520086",
					"description": "Leo3 Test",
					"email": "leotest3@gmail.com"
				},
				"messages": {
					"resultCode": "Ok",
					"message": [{
						"code": "I00001",
						"text": "Successful."
					}]
				}
			}`))
		}))
		defer server.Close()

		svc := &service{
			apiLoginID:     apiLoginID,
			transactionKey: transactionKey,
			endpoint:       server.URL,
			httpClient:     &http.Client{},
			logger:         zap.NewExample().Sugar(),
		}

		req := entities.GetPaymentProfilesRequest{
			ProfileID: profileID,
		}
		res, err := svc.GetCustomerPaymentProfiles(ctx, req)

		assert.NoError(t, err)
		assert.Len(t, res.PaymentProfiles, 1)
		assert.Equal(t, "535672235", res.PaymentProfiles[0].ID)
		assert.Equal(t, "XXXX8888", res.PaymentProfiles[0].CardNumber)
		assert.Equal(t, "Visa", res.PaymentProfiles[0].CardType)
		assert.Equal(t, "2025-12", res.PaymentProfiles[0].ExpirationDate)
	})

	t.Run("No Payment Profiles", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"profile": {
					"profileType": "regular",
					"customerProfileId": "523542388", 
					"merchantCustomerId": "4d6aad2a-c3f1-4888-a",
					"description": "Test from testing",
					"email": "testing2@example.com"
				},
				"messages": {
					"resultCode": "Ok",
					"message": [{
						"code": "I00001",
						"text": "Successful."
					}]
				}
			}`))
		}))
		defer server.Close()

		svc := &service{
			apiLoginID:     apiLoginID,
			transactionKey: transactionKey,
			endpoint:       server.URL,
			httpClient:     &http.Client{},
			logger:         zap.NewExample().Sugar(),
		}

		req := entities.GetPaymentProfilesRequest{
			ProfileID: profileID,
		}

		res, err := svc.GetCustomerPaymentProfiles(ctx, req)

		assert.NoError(t, err)
		assert.Empty(t, res.PaymentProfiles)
	})

	t.Run("Profile Not Found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"messages": {
					"resultCode": "Error",
					"message": [
						{
							"code": "E00013",
							"text": "Customer Profile ID is invalid."
						}
					]
				}
			}`))
		}))
		defer server.Close()

		svc := &service{
			apiLoginID:     apiLoginID,
			transactionKey: transactionKey,
			endpoint:       server.URL,
			httpClient:     &http.Client{},
			logger:         zap.NewExample().Sugar(),
		}

		req := entities.GetPaymentProfilesRequest{
			ProfileID: profileID,
		}

		res, err := svc.GetCustomerPaymentProfiles(ctx, req)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "Customer Profile ID is invalid.")
		assert.Empty(t, res.PaymentProfiles)
	})

	t.Run("HTTP Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		svc := &service{
			apiLoginID:     apiLoginID,
			transactionKey: transactionKey,
			endpoint:       server.URL,
			httpClient:     &http.Client{},
			logger:         zap.NewExample().Sugar(),
		}

		req := entities.GetPaymentProfilesRequest{
			ProfileID: profileID,
		}

		res, err := svc.GetCustomerPaymentProfiles(ctx, req)

		assert.Error(t, err)
		assert.Empty(t, res.PaymentProfiles)
	})
}

func TestCreatePaymentTransaction(t *testing.T) {
	ctx := context.TODO()
	profileID := "test-profile-id"
	apiLoginID := "test-login"
	transactionKey := "test-key"
	expectedID := "80041310709"

	t.Run("Success: Approved", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var requestBody CreateTransactionRequest
			err := json.NewDecoder(r.Body).Decode(&requestBody)
			assert.NoError(t, err)

			assert.Equal(t, apiLoginID, requestBody.Data.MerchantAuthentication.Name)
			assert.Equal(t, transactionKey, requestBody.Data.MerchantAuthentication.TransactionKey)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(fmt.Appendf(nil, `{
				"transactionResponse": {
					"responseCode": "1",					
					"transId": "%s",
					"accountNumber": "XXXX1111",
					"accountType": "Visa",
					"messages": [{
						"code": "1",
						"description": "This transaction has been approved."
					}]
				},
				"messages": {
					"resultCode": "Ok",
					"message": [{
						"code": "I00001",
						"text": "Successful."
					}]
				}
			}`, expectedID))
		}))
		defer server.Close()

		svc := &service{
			apiLoginID:     apiLoginID,
			transactionKey: transactionKey,
			endpoint:       server.URL,
			httpClient:     &http.Client{},
			logger:         zap.NewExample().Sugar(),
		}

		req := entities.CreatePaymentTransactionRequest{
			ProfileID:    profileID,
			Amount:       decimal.NewFromInt(100),
			PaymentNonce: "1234567890",
		}

		res, err := svc.CreatePaymentTransaction(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, expectedID, res.ID)
		assert.Equal(t, AuthorizeNetStatusApproved, res.Status)
	})

	t.Run("Success: Declined", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"transactionResponse": {
					"responseCode": "2",
					"transId": "80041310908",
					"accountNumber": "XXXX1111",
					"accountType": "Visa",
					"errors": [{
						"errorCode": "2",
						"errorText": "This transaction has been declined."
					}]
				},
				"messages": {
					"resultCode": "Ok",
					"message": [{
						"code": "I00001",
						"text": "Successful."
					}]
				}
			}`))
		}))
		defer server.Close()

		svc := &service{
			apiLoginID:     apiLoginID,
			transactionKey: transactionKey,
			endpoint:       server.URL,
			httpClient:     &http.Client{},
			logger:         zap.NewExample().Sugar(),
		}

		req := entities.CreatePaymentTransactionRequest{
			ProfileID:    profileID,
			Amount:       decimal.NewFromInt(100),
			PaymentNonce: "1234567890",
		}

		res, err := svc.CreatePaymentTransaction(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "80041310908", res.ID)
		assert.Equal(t, AuthorizeNetStatusDeclined, res.Status)
	})

	t.Run("API Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"transactionResponse": {
					"SupplementalDataQualificationIndicator": 0
				},
				"messages": {
					"resultCode": "Error",
					"message": [{
						"code": "E00114",
						"text": "Invalid OTS Token."
					}]
				}
			}`))
		}))
		defer server.Close()

		svc := &service{
			apiLoginID:     apiLoginID,
			transactionKey: transactionKey,
			endpoint:       server.URL,
			httpClient:     &http.Client{},
			logger:         zap.NewExample().Sugar(),
		}

		req := entities.CreatePaymentTransactionRequest{
			ProfileID:    profileID,
			Amount:       decimal.NewFromInt(100),
			PaymentNonce: "1234567890",
		}

		res, err := svc.CreatePaymentTransaction(ctx, req)

		assert.Error(t, err)
		assert.Empty(t, res)
	})

	t.Run("HTTP Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		svc := &service{
			apiLoginID:     apiLoginID,
			transactionKey: transactionKey,
			endpoint:       server.URL,
			httpClient:     &http.Client{},
			logger:         zap.NewExample().Sugar(),
		}

		req := entities.CreatePaymentTransactionRequest{
			ProfileID:    profileID,
			Amount:       decimal.NewFromInt(100),
			PaymentNonce: "1234567890",
		}

		res, err := svc.CreatePaymentTransaction(ctx, req)

		assert.Error(t, err)
		assert.Empty(t, res)
	})

}

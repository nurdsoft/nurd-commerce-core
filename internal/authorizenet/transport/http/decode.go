package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/internal/authorizenet/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"

	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"

	goKitHTTPTransport "github.com/go-kit/kit/transport/http"
	httpError "github.com/nurdsoft/nurd-commerce-core/shared/errors/http"
	"github.com/pkg/errors"
)

func decodeBodyFromRequest[T any](req *T, r *http.Request) error {
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return errors.WithMessage(httpError.ErrBadRequestBody, err.Error())
	}

	defer r.Body.Close()

	return nil
}

func decodeGetPaymentProfiles(c context.Context, _ *http.Request) (interface{}, error) {
	customerIDStr := sharedMeta.XCustomerID(c)

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "Customer ID is not valid")
	}

	if customerID == uuid.Nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "Customer ID not found in context")
	}
	return nil, nil
}

func decodeCreatePaymentProfileRequest(c context.Context, r *http.Request) (interface{}, error) {
	customerIDStr := sharedMeta.XCustomerID(c)

	if err := validateCustomerID(customerIDStr); err != nil {
		return nil, err
	}

	var req entities.CreatePaymentProfileRequest
	err := decodeBodyFromRequest(&req, r)
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "unable to read request body")
	}

	return req, nil
}

// NewDecodeWebhookRequest returns a decode function with the signature key injected
func NewDecodeWebhookRequest(signatureKey string) goKitHTTPTransport.DecodeRequestFunc {
	return func(_ context.Context, r *http.Request) (interface{}, error) {
		signature := r.Header.Get("X-Anet-Signature")
		if signature == "" {
			return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "missing X-Anet-Signature header")
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "unable to read request body")
		}

		err = validateSignature(signature, bodyBytes, signatureKey)
		if err != nil {
			return nil, err
		}

		var req entities.WebhookRequest
		err = json.Unmarshal(bodyBytes, &req)
		if err != nil {
			return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "unable to decode request body")
		}

		return req, nil
	}
}

func validateCustomerID(customerIDStr string) error {
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return moduleErrors.NewAPIError("VALIDATION_ERROR", "Customer ID is not valid")
	}

	if customerID == uuid.Nil {
		return moduleErrors.NewAPIError("VALIDATION_ERROR", "Customer ID not found in context")
	}

	return nil
}

func validateSignature(signature string, body []byte, signatureKey string) error {
	parts := strings.SplitN(signature, "=", 2)
	if len(parts) != 2 || parts[0] != "sha512" {
		return moduleErrors.NewAPIError("VALIDATION_ERROR", "invalid signature format")
	}

	expectedSignature, err := hex.DecodeString(parts[1])
	if err != nil {
		return moduleErrors.NewAPIError("VALIDATION_ERROR", "invalid signature format")
	}

	h := hmac.New(sha512.New, []byte(signatureKey))
	h.Write(body)
	computedSignature := h.Sum(nil)

	if !hmac.Equal(expectedSignature, computedSignature) {
		return moduleErrors.NewAPIError("VALIDATION_ERROR", "invalid webhook signature")
	}

	return nil
}

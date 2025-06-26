package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nurdsoft/nurd-commerce-core/internal/stripe/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	httpError "github.com/nurdsoft/nurd-commerce-core/shared/errors/http"
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"github.com/pkg/errors"
)

type RequestBodyType interface {
	entities.StripeRefundRequestBody
}

func decodeBodyFromRequest[T RequestBodyType](req *T, r *http.Request) error {
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return errors.WithMessage(httpError.ErrBadRequestBody, err.Error())
	}

	defer r.Body.Close()

	return nil
}

func decodeStripeGetPaymentMethods(c context.Context, _ *http.Request) (interface{}, error) {
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

func decodeStripeGetPaymentMethod(c context.Context, r *http.Request) (interface{}, error) {
	customerIDStr := sharedMeta.XCustomerID(c)

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "Customer ID is not valid")
	}

	if customerID == uuid.Nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "Customer ID not found in context")
	}

	params := mux.Vars(r)
	paymentMethodId := params["payment_method_id"]
	if paymentMethodId == "" {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "payment_method_id is not valid")
	}

	return &entities.StripeGetPaymentMethodRequest{
		PaymentMethodId: paymentMethodId,
	}, nil
}

func decodeStripeWebhookRequest(_ context.Context, r *http.Request) (interface{}, error) {
	signature := r.Header.Get("Stripe-Signature")
	if signature == "" {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "missing Stripe-Signature header")
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "unable to read request body")
	}
	defer r.Body.Close()

	return &entities.StripeWebhookRequest{
		Payload:   payload,
		Signature: signature,
	}, nil
}

func decodeStripeRefundRequest(c context.Context, r *http.Request) (interface{}, error) {
	customerIDStr := sharedMeta.XCustomerID(c)

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "Customer ID is not valid")
	}

	if customerID == uuid.Nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "Customer ID not found in context")
	}

	req := &entities.StripeRefundRequestBody{}
	if err := decodeBodyFromRequest(req, r); err != nil {
		return nil, err
	}
	params := mux.Vars(r)
	paymentIntentId := params["payment_intent_id"]
	if paymentIntentId == "" {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "PaymentIntentId is required")
	}

	return &entities.StripeRefundRequest{
		PaymentIntentId: paymentIntentId,
		Body:            req,
	}, nil
}

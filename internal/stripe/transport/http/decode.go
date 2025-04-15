package http

import (
	"context"
	"encoding/json"
	"github.com/nurdsoft/nurd-commerce-core/internal/stripe/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"github.com/google/uuid"
	"io"
	"net/http"

	httpError "github.com/nurdsoft/nurd-commerce-core/shared/errors/http"
	"github.com/pkg/errors"
)

type RequestBodyType interface {
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

package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"

	"github.com/nurdsoft/nurd-commerce-core/internal/customer/entities"
	httpError "github.com/nurdsoft/nurd-commerce-core/shared/errors/http"
	"github.com/pkg/errors"
)

type RequestBodyType interface {
	entities.CreateCustomerRequestBody |
		entities.UpdateCustomerRequestBody
}

func decodeBodyFromRequest[T RequestBodyType](req *T, r *http.Request) error {
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return errors.WithMessage(httpError.ErrBadRequestBody, err.Error())
	}

	defer r.Body.Close()

	return nil
}

func decodeCreateCustomerRequest(_ context.Context, r *http.Request) (interface{}, error) {
	reqBody := &entities.CreateCustomerRequestBody{}
	err := decodeBodyFromRequest(reqBody, r)
	if err != nil {
		return nil, err
	}

	if reqBody.FirstName == "" {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "First name is required")
	}

	if reqBody.Email == "" {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "Email is required")
	}

	return &entities.CreateCustomerRequest{
		Data: reqBody,
	}, nil
}

func decodeGetCustomerRequest(c context.Context, _ *http.Request) (interface{}, error) {
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

func decodeUpdateCustomerRequest(c context.Context, r *http.Request) (interface{}, error) {
	customerIDStr := sharedMeta.XCustomerID(c)

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "Customer ID is not valid")
	}

	if customerID == uuid.Nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "Customer ID not found in context")
	}

	reqBody := &entities.UpdateCustomerRequestBody{}
	err = decodeBodyFromRequest(reqBody, r)
	if err != nil {
		return nil, err
	}

	return &entities.UpdateCustomerRequest{
		Data: reqBody,
	}, nil
}

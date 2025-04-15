package http

import (
	"context"
	"encoding/json"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/internal/cart/errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	httpError "github.com/nurdsoft/nurd-commerce-core/shared/errors/http"
	"github.com/pkg/errors"
)

type RequestBodyType interface {
	entities.UpdateCartItemRequestBody |
		entities.GetTaxRateRequestBody |
		entities.GetShippingRateRequestBody
}

func decodeBodyFromRequest[T RequestBodyType](req *T, r *http.Request) error {
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return errors.WithMessage(httpError.ErrBadRequestBody, err.Error())
	}

	defer r.Body.Close()

	return nil
}

func decodeUpdateCartItemRequest(_ context.Context, r *http.Request) (interface{}, error) {
	reqBody := &entities.UpdateCartItemRequestBody{}
	err := decodeBodyFromRequest(reqBody, r)
	if err != nil {
		return nil, err
	}

	return &entities.UpdateCartItemRequest{
		Item: reqBody,
	}, nil
}

func decodeGetCartItemsRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeRemoveCartItemRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	itemId, ok := vars["item_id"]
	if !ok {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "item_id not found in path")
	}

	return itemId, nil
}

func decodeClearCartItemsRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeGetShippingRateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	reqBody := &entities.GetShippingRateRequestBody{}
	err := decodeBodyFromRequest(reqBody, r)
	if err != nil {
		return nil, err
	}

	return &entities.GetShippingRateRequest{
		Body: reqBody,
	}, nil
}

func decodeGetTaxRateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	reqBody := &entities.GetTaxRateRequestBody{}
	err := decodeBodyFromRequest(reqBody, r)
	if err != nil {
		return nil, err
	}

	return &entities.GetTaxRateRequest{
		Body: reqBody,
	}, nil
}

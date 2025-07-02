package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/internal/orders/errors"
	httpError "github.com/nurdsoft/nurd-commerce-core/shared/errors/http"
	"github.com/pkg/errors"
)

type RequestBodyType interface {
	entities.CreateOrderRequestBody |
		entities.UpdateOrderRequestBody |
		entities.RefundOrderRequestBody
}

func decodeBodyFromRequest[T RequestBodyType](req *T, r *http.Request) error {
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return errors.WithMessage(httpError.ErrBadRequestBody, err.Error())
	}

	defer r.Body.Close()

	return nil
}

func decodeCreateOrderRequest(_ context.Context, r *http.Request) (interface{}, error) {
	reqBody := &entities.CreateOrderRequestBody{}
	err := decodeBodyFromRequest(reqBody, r)
	if err != nil {
		return nil, err
	}

	return &entities.CreateOrderRequest{
		Body: reqBody,
	}, nil
}

func decodeListOrdersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "limit is required")
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "invalid limit")
	}

	var includeItems bool

	includeItemsStr := r.URL.Query().Get("include_items")
	if includeItemsStr == "" {
		includeItems = false
	} else {
		includeItems, err = strconv.ParseBool(includeItemsStr)
		if err != nil {
			return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "invalid value for include_items")
		}
	}

	return &entities.ListOrdersRequest{
		Limit:        limit,
		Cursor:       r.URL.Query().Get("cursor"),
		IncludeItems: includeItems,
	}, nil
}

func decodeGetOrderRequest(_ context.Context, r *http.Request) (interface{}, error) {
	params := mux.Vars(r)
	orderID, err := uuid.Parse(params["order_id"])
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "invalid order_id")
	}

	return &entities.GetOrderRequest{
		OrderID: orderID,
	}, nil
}

func decodeCancelOrderRequest(_ context.Context, r *http.Request) (interface{}, error) {
	params := mux.Vars(r)
	orderID, err := uuid.Parse(params["order_id"])
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "invalid order_id")
	}

	return &entities.CancelOrderRequest{
		OrderID: orderID,
	}, nil
}

func decodeUpdateOrderRequest(_ context.Context, r *http.Request) (interface{}, error) {
	params := mux.Vars(r)
	orderReference := params["order_reference"]
	if orderReference == "" {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "invalid order_reference")
	}

	reqBody := &entities.UpdateOrderRequestBody{}
	err := decodeBodyFromRequest(reqBody, r)
	if err != nil {
		return nil, err
	}

	return &entities.UpdateOrderRequest{
		OrderReference: orderReference,
		Body:           reqBody,
	}, nil
}

func decodeRefundOrderRequest(_ context.Context, r *http.Request) (interface{}, error) {
	params := mux.Vars(r)
	orderReference := params["order_reference"]
	if orderReference == "" {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "invalid order_reference")
	}

	reqBody := &entities.RefundOrderRequestBody{}
	err := decodeBodyFromRequest(reqBody, r)
	if err != nil {
		return nil, err
	}

	return &entities.RefundOrderRequest{
		OrderReference: orderReference,
		Body:           reqBody,
	}, nil
}

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/internal/wishlist/errors"
	httpError "github.com/nurdsoft/nurd-commerce-core/shared/errors/http"
	"github.com/pkg/errors"
)

type RequestBodyType interface {
	entities.AddToWishlistRequestBody |
		entities.GetWishlistProductTimestampsRequestBody
}

func decodeBodyFromRequest[T RequestBodyType](req *T, r *http.Request) error {
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return errors.WithMessage(httpError.ErrBadRequestBody, err.Error())
	}

	defer r.Body.Close()

	return nil
}

func decodeAddToWishlistRequest(_ context.Context, r *http.Request) (interface{}, error) {
	reqBody := &entities.AddToWishlistRequestBody{}
	err := decodeBodyFromRequest(reqBody, r)
	if err != nil {
		return nil, err
	}

	return &entities.AddToWishlistRequest{
		Body: reqBody,
	}, nil
}

func decodeRemoveFromWishlistRequest(_ context.Context, r *http.Request) (interface{}, error) {
	params := mux.Vars(r)
	productID, err := uuid.Parse(params["product_id"])
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "product_id is not valid")
	}

	return &entities.RemoveFromWishlistRequest{
		ProductID: productID,
	}, nil
}

func decodeGetWishlistRequest(_ context.Context, r *http.Request) (interface{}, error) {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "limit is required")
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "invalid limit")
	}

	return &entities.GetWishlistRequest{
		Limit:  limit,
		Cursor: r.URL.Query().Get("cursor"),
	}, nil
}

func decodeGetMoreFromWishlistRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var limit int
	var err error
	limitStr := r.URL.Query().Get("limit")

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "invalid limit")
		}
	}

	return &entities.GetMoreFromWishlistRequest{
		Limit:  limit,
		Cursor: r.URL.Query().Get("cursor"),
	}, nil
}

func decodeGetWishlistProductTimestampsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	reqBody := &entities.GetWishlistProductTimestampsRequestBody{}
	err := decodeBodyFromRequest(reqBody, r)
	if err != nil {
		return nil, err
	}

	return &entities.GetWishlistProductTimestampsRequest{
		Body: reqBody,
	}, nil
}

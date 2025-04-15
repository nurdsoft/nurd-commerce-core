package http

import (
	"context"
	"encoding/json"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/nurdsoft/nurd-commerce-core/internal/product/entities"
	httpError "github.com/nurdsoft/nurd-commerce-core/shared/errors/http"
	"github.com/pkg/errors"
)

type RequestBodyType interface {
	entities.CreateProductRequestBody | entities.CreateProductVariantRequestBody
}

func decodeBodyFromRequest[T RequestBodyType](req *T, r *http.Request) error {
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return errors.WithMessage(httpError.ErrBadRequestBody, err.Error())
	}

	defer r.Body.Close()

	return nil
}

func decodeCreateProductRequest(_ context.Context, r *http.Request) (interface{}, error) {
	reqBody := &entities.CreateProductRequestBody{}
	err := decodeBodyFromRequest(reqBody, r)
	if err != nil {
		return nil, err
	}

	if reqBody.Name == "" {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "Name is required")
	}

	return &entities.CreateProductRequest{
		Data: reqBody,
	}, nil
}

func decodeGetProductRequest(_ context.Context, r *http.Request) (interface{}, error) {
	params := mux.Vars(r)
	productID, err := uuid.Parse(params["product_id"])
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "product_id is not valid")
	}

	return &entities.GetProductRequest{
		ProductID: productID,
	}, nil
}

func decodeGetProductVariantRequest(_ context.Context, r *http.Request) (interface{}, error) {
	params := mux.Vars(r)
	sku := params["sku"]
	if sku == "" {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "SKU is not valid")
	}

	return &entities.GetProductVariantRequest{
		SKU: sku,
	}, nil
}

func decodeCreateProductVariantRequest(_ context.Context, r *http.Request) (interface{}, error) {
	params := mux.Vars(r)
	productID, err := uuid.Parse(params["product_id"])
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "product_id is not valid")
	}

	reqBody := &entities.CreateProductVariantRequestBody{}
	err = decodeBodyFromRequest(reqBody, r)
	if err != nil {
		return nil, err
	}

	if reqBody.SKU == "" {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "SKU is required")
	}

	if reqBody.Name == "" {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "Name is required")
	}

	return &entities.CreateProductVariantRequest{
		ProductID: productID,
		Data: reqBody,
	}, nil
}

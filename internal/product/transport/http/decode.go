package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"

	"github.com/nurdsoft/nurd-commerce-core/internal/product/entities"
	httpError "github.com/nurdsoft/nurd-commerce-core/shared/errors/http"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
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
		Data:      reqBody,
	}, nil
}

var validSortFields = map[string]struct{}{
	"created_at": {},
	"updated_at": {},
	"name":       {},
	"price":      {},
}

func decodeListProductVariantsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	query := r.URL.Query()

	// Parse pagination parameters
	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := 10
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if parsed, err := strconv.Atoi(pageSizeStr); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	// Parse search parameter
	var search *string
	if searchStr := query.Get("search"); searchStr != "" {
		search = &searchStr
	}

	// Parse price filters
	var minPrice, maxPrice *decimal.Decimal
	if minPriceStr := query.Get("min_price"); minPriceStr != "" {
		if parsed, err := decimal.NewFromString(minPriceStr); err == nil {
			minPrice = &parsed
		}
	}
	if maxPriceStr := query.Get("max_price"); maxPriceStr != "" {
		if parsed, err := decimal.NewFromString(maxPriceStr); err == nil {
			maxPrice = &parsed
		}
	}

	// Parse sort parameters
	var sortBy *string
	if sortByStr := query.Get("sort_by"); sortByStr != "" {
		if _, ok := validSortFields[sortByStr]; ok {
			sortBy = &sortByStr
		}
	}

	var sortOrder *string
	if sortOrderStr := query.Get("sort_order"); sortOrderStr != "" {
		if sortOrderStr == "asc" || sortOrderStr == "desc" {
			sortOrder = &sortOrderStr
		}
	}

	// Parse attributes filter
	attributes := make(map[string]string)
	for key, values := range query {
		if strings.HasPrefix(key, "attributes[") && strings.HasSuffix(key, "]") {
			attrKey := strings.TrimPrefix(key, "attributes[")
			attrKey = strings.TrimSuffix(attrKey, "]")
			if len(values) > 0 {
				attributes[attrKey] = values[0]
			}
		}
	}

	return &entities.ListProductVariantsRequest{
		Page:       page,
		PageSize:   pageSize,
		Search:     search,
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		SortBy:     sortBy,
		SortOrder:  sortOrder,
		Attributes: attributes,
	}, nil
}

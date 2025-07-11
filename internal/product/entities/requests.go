package entities

import (
	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/shared/json"
	"github.com/shopspring/decimal"
)

// swagger:parameters products CreateProductRequest
type CreateProductRequest struct {
	// Product data to be created
	//
	// required: true
	// in:body
	Data *CreateProductRequestBody
}

type CreateProductRequestBody struct {
	ID          *uuid.UUID `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	ImageURL    *string    `json:"image_url"`
	Attributes  *json.JSON `json:"attributes"`
}

// swagger:parameters products UpdateProductRequest
type UpdateProductRequest struct {
	// Product ID to be fetched
	//
	// in:path
	ProductID uuid.UUID `json:"product_id"`
	// Product data to be created
	//
	// required: true
	// in:body
	Data *UpdateProductRequestBody
}

type UpdateProductRequestBody struct {
	SalesforceID               string `json:"salesforce_id"`
	SalesforcePricebookEntryId string `json:"salesforce_pricebook_entry_id"`
}

// swagger:parameters products GetProductRequest
type GetProductRequest struct {
	// Product ID to be fetched
	//
	// in:path
	ProductID uuid.UUID `json:"product_id"`
}

// swagger:parameters products CreateProductVariantRequest
type CreateProductVariantRequest struct {
	// Product ID to be created the variant for
	//
	// in:path
	ProductID uuid.UUID `json:"product_id"`
	// Product variant data to be created
	//
	// required: true
	// in:body
	Data *CreateProductVariantRequestBody
}

type CreateProductVariantRequestBody struct {
	SKU           string           `json:"sku"`
	Name          string           `json:"name"`
	Description   *string          `json:"description"`
	ImageURL      *string          `json:"image_url"`
	Price         decimal.Decimal  `json:"price"`
	Currency      string           `json:"currency"`
	Length        *decimal.Decimal `json:"length"`
	Width         *decimal.Decimal `json:"width"`
	Height        *decimal.Decimal `json:"height"`
	Weight        *decimal.Decimal `json:"weight"`
	Attributes    *json.JSON       `json:"attributes"`
	StripeTaxCode *string          `json:"stripe_tax_code"`
}

// swagger:parameters products GetProductVariantRequest
type GetProductVariantRequest struct {
	// Product variant SKU to be fetched
	//
	// in:path
	SKU string `json:"sku"`
}

// swagger:parameters products GetProductsRequest
type GetProductsRequest struct {
	// Product ID to be fetched
	//
	// in:path
	ProductIDs []uuid.UUID `json:"product_ids"`
}

// swagger:parameters products ListProductVariantsRequest
type ListProductVariantsRequest struct {
	// Page number for pagination. Default: 1
	//
	// in:query
	Page int `json:"page"`
	// Page size for pagination. Default: 10
	//
	// in:query
	PageSize int `json:"page_size"`
	// Search term for name/description (optional)
	//
	// in:query
	Search *string `json:"search"`
	// Minimum price filter (optional)
	//
	// in:query
	MinPrice *decimal.Decimal `json:"min_price"`
	// Maximum price filter (optional)
	//
	// in:query
	MaxPrice *decimal.Decimal `json:"max_price"`
	// Sort field (optional), Default: created_at
	//
	// in:query
	SortBy *string `json:"sort_by"`
	// Sort direction (optional), Default: desc
	//
	// in:query
	SortOrder *string `json:"sort_order"`
	// JSON attributes filter (optional) - format: attributes[key]=value
	//
	// swagger:ignore
	Attributes map[string]string `json:"attributes"`
}

// swagger:model ListProductVariantsResponse
type ListProductVariantsResponse struct {
	Data       []ProductVariant `json:"data"`
	Pagination PaginationMeta   `json:"pagination"`
}

// swagger:model PaginationMeta
type PaginationMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

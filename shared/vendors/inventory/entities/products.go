package entities

import (
	"encoding/json"
	"time"
)

type PaginationMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type ListProductsRequest struct {
	Search   string
	Page     int
	PageSize int
}

type ListProductsResponse struct {
	Data       []ProductResponse `json:"data"`
	Pagination PaginationMeta    `json:"pagination"`
}

type ProductResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	ImageURL *string `json:"image_url"`
}

type Product struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description *string          `json:"description,omitempty"`
	ImageURL    *string          `json:"image_url"`
	Attributes  *json.RawMessage `json:"attributes,omitempty"`
	Variants    []ProductVariant `json:"variants"`
	CreatedAt   time.Time        `json:"created_at,omitzero"`
	UpdatedAt   *time.Time       `json:"updated_at,omitempty"`
}

type ProductVariant struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	SKU        string           `json:"sku"`
	ImageURL   *string          `json:"image_url"`
	Attributes *json.RawMessage `json:"attributes"`
}

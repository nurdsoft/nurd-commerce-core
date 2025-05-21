package entities

import (
	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/shared/json"
)

// swagger:parameters wishlist AddToWishlistRequest
type AddToWishlistRequest struct {
	// Products to be added to wishlist
	//
	// in:body
	Body *AddToWishlistRequestBody
}

type AddToWishlistRequestBody struct {
	Products []Product `json:"products"`
}

type Product struct {
	ProductID   uuid.UUID    `json:"product_id"`
	ProductData *ProductData `json:"data"`
}

type ProductData struct {
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	ImageURL    *string    `json:"image_url,omitempty"`
	Attributes  *json.JSON `json:"attributes"`
}

// swagger:parameters wishlist RemoveFromWishlistRequest
type RemoveFromWishlistRequest struct {
	// Product ID to be removed
	//
	// in:path
	ProductID uuid.UUID `json:"product_id"`
}

// swagger:parameters wishlist GetWishlistRequest
type GetWishlistRequest struct {
	// Limit of items to return
	//
	// required: true
	// in:query
	Limit int `json:"limit"`
	// Cursor to paginate orders
	//
	// in:query
	Cursor string `json:"cursor"`
}

type BulkRemoveFromWishlistRequest struct {
	CustomerID uuid.UUID   `json:"customer_id"`
	ProductIDs []uuid.UUID `json:"product_ids"`
}

// swagger:parameters wishlist GetMoreFromWishlistRequest
type GetMoreFromWishlistRequest struct {
	// Limit of orders to return
	//
	// in:query
	Limit int `json:"limit"`
	// Cursor to paginate orders
	//
	// in:query
	Cursor string `json:"cursor"`
}

// swagger:parameters wishlist GetWishlistProductTimestampsRequest
type GetWishlistProductTimestampsRequest struct {
	// Products to get timestamps for
	//
	// in:body
	Body *GetWishlistProductTimestampsRequestBody
}

type GetWishlistProductTimestampsRequestBody struct {
	ProductIDs []uuid.UUID `json:"product_ids"`
}

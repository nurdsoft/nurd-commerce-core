package entities

import "time"

// swagger:model GetWishlistResponse
type GetWishlistResponse struct {
	Items      []*WishlistItem `json:"items"`
	NextCursor string          `json:"next_cursor"`
	Total      int64           `json:"total"` // Total number of items in the wishlist
}

// swagger:model GetWishlistProductTimestampsResponse
type GetWishlistProductTimestampsResponse struct {
	// Map of product ID to wishlist timestamp
	Timestamps map[string]time.Time `json:"timestamps"`
}

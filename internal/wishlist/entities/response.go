package entities

// swagger:model GetWishlistResponse
type GetWishlistResponse struct {
	Items      []*WishlistItem `json:"items"`
	NextCursor string          `json:"next_cursor"`
}

package entities

import (
	"time"

	"github.com/google/uuid"
)

type WishlistItem struct {
	Id         uuid.UUID `json:"-" db:"id"`
	CustomerID uuid.UUID `json:"-" db:"customer_id"`
	ProductID  uuid.UUID `json:"product_id" db:"product_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

func (w *WishlistItem) TableName() string {
	return "wishlist_items"
}

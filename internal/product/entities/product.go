package entities

import (
	"time"

	"github.com/nurdsoft/nurd-commerce-core/shared/json"

	"github.com/google/uuid"
)

// swagger:model GetProductResponse
type Product struct {
	ID                         uuid.UUID  `json:"id" db:"id"`
	Name                       string     `json:"name" db:"name"`
	Description                *string    `json:"description" db:"description"`
	ImageURL                   *string    `json:"image_url" db:"image_url"`
	Attributes                 *json.JSON `json:"attributes" db:"attributes"`
	SalesforceID               *string    `json:"-" db:"salesforce_id"`
	SalesforcePricebookEntryId *string    `json:"-" db:"salesforce_pricebook_entry_id"`
	CreatedAt                  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                  *time.Time `json:"updated_at" db:"updated_at"`
}

func (u *Product) TableName() string {
	return "products"
}

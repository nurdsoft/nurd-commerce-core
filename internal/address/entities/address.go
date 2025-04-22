package entities

import (
	"time"

	"github.com/google/uuid"
)

// swagger:model GetAddressResponse
type Address struct {
	ID           uuid.UUID `json:"id" gorm:"column:id"`
	CustomerID   uuid.UUID `json:"-" gorm:"column:customer_id"`
	FullName     string    `json:"full_name" gorm:"column:full_name"`
	Address      string    `json:"address" gorm:"column:address"`
	Apartment    *string   `json:"apartment" gorm:"column:apartment"`
	City         *string   `json:"city" gorm:"column:city"`
	PhoneNumber  *string   `json:"phone_number" gorm:"column:phone_number"`
	StateCode    string    `json:"state_code" gorm:"column:state_code"`
	CountryCode  string    `json:"country_code" gorm:"column:country_code"`
	PostalCode   string    `json:"postal_code" gorm:"column:postal_code"`
	IsDefault    bool      `json:"is_default" gorm:"column:is_default"`
	SalesforceID *string   `json:"-" gorm:"column:salesforce_id"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (Address) TableName() string {
	return "addresses"
}

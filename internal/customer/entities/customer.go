package entities

import (
	"time"

	"github.com/google/uuid"
)

// swagger:model GetCustomerResponse
type Customer struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Email          string     `json:"email" db:"email"`
	FirstName      string     `json:"first_name" db:"first_name"`
	LastName       *string    `json:"last_name" db:"last_name"`
	PhoneNumber    *string    `json:"phone_number" db:"phone_number"`
	SalesforceID   *string    `json:"salesforce_id" db:"salesforce_id"`
	StripeID       *string    `json:"-" db:"stripe_id"`
	AuthorizeNetID *string    `json:"-" gorm:"column:authorizenet_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at" db:"updated_at"`
}

func (u *Customer) TableName() string {
	return "customers"
}

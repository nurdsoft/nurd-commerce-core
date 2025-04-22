package entities

import "github.com/google/uuid"

// swagger:parameters customers CreateCustomerRequest
type CreateCustomerRequest struct {
	// Customer data to be created
	//
	// required: true
	// in:body
	Data *CreateCustomerRequestBody
}

type CreateCustomerRequestBody struct {
	ID           *uuid.UUID `json:"id,omitempty"`
	Email        string     `json:"email"`
	FirstName    string     `json:"first_name"`
	LastName     *string    `json:"last_name,omitempty"`
	PhoneNumber  *string    `json:"phone_number,omitempty"`
}

// swagger:parameters customers UpdateCustomerRequest
type UpdateCustomerRequest struct {
	// Customer data to be updated
	//
	// required: true
	// in:body
	Data *UpdateCustomerRequestBody
}

type UpdateCustomerRequestBody struct {
	FirstName    string  `json:"first_name"`
	LastName     *string `json:"last_name,omitempty"`
	PhoneNumber  *string `json:"phone_number,omitempty"`
}

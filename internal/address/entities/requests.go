package entities

import "github.com/google/uuid"

// swagger:parameters addresses AddAddressRequest
type AddAddressRequest struct {
	// Address to be added
	//
	// in:body
	Address *AddressRequestBody `json:"address"`
}

type AddressRequestBody struct {
	FullName    string  `json:"full_name"`
	Address     string  `json:"address"`
	Apartment   *string `json:"apartment,omitempty"`
	City        *string `json:"city,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	StateCode   string  `json:"state_code"`
	CountryCode string  `json:"country_code"`
	PostalCode  string  `json:"postal_code"`
	IsDefault   bool    `json:"is_default"`
}

// swagger:parameters addresses UpdateAddressRequest
type UpdateAddressRequest struct {
	// Address UUID to be updated
	//
	// in:path
	AddressID uuid.UUID `json:"address_id"`
	// Address to be updated
	//
	// in:body
	Address *AddressRequestBody `json:"address"`
}

// swagger:parameters addresses DeleteAddressRequest
type DeleteAddressRequest struct {
	// Address UUID to be deleted
	//
	// in:path
	AddressID uuid.UUID `json:"address_id"`
}

// swagger:parameters addresses GetAddressRequest
type GetAddressRequest struct {
	// Address UUID to be fetched
	//
	// in:path
	AddressID uuid.UUID `json:"address_id"`
}

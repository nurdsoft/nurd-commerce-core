package entities

import (
	"github.com/shopspring/decimal"
	"time"
)

type Address struct {
	FullName    string
	Address     string
	City        string
	StateCode   string
	PostalCode  string
	CountryCode string
}

type Shipment struct {
	Origin      Address
	Destination Address
	Dimensions  Dimensions
}

type Dimensions struct {
	Length decimal.Decimal
	Width  decimal.Decimal
	Height decimal.Decimal
	Weight decimal.Decimal
}

type ValidationResult struct {
	Valid   bool
	Message string
}

type ShippingRate struct {
	Amount                decimal.Decimal
	Currency              string
	CarrierName           string
	CarrierCode           string
	ServiceType           string
	ServiceCode           string
	EstimatedDeliveryDate time.Time
	BusinessDaysInTransit string
	CreatedAt             time.Time
}

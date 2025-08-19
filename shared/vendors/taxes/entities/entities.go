package entities

import (
	"github.com/nurdsoft/nurd-commerce-core/shared/json"
	"github.com/shopspring/decimal"
)

type CalculateTaxRequest struct {
	ShippingAmount decimal.Decimal
	FromAddress    *Address
	ToAddress      Address
	TaxItems       []TaxItem
}

type TaxItem struct {
	Price     decimal.Decimal
	Quantity  int
	Reference string
	TaxCode   string
}

type Address struct {
	Line1      string
	City       string
	State      string
	PostalCode string
	Country    string
}

type CalculateTaxResponse struct {
	Tax         decimal.Decimal
	TotalAmount decimal.Decimal
	Currency    string
	Breakdown   json.JSON
}

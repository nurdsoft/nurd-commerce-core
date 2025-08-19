package entities

import (
	"github.com/nurdsoft/nurd-commerce-core/shared/json"
	"github.com/shopspring/decimal"
)

type CalculateTaxRequest struct {
	// Shipping amount, in the smallest currency unit.
	// https://docs.stripe.com/currencies#minor-units
	ShippingAmount decimal.Decimal
	// Address of warehouse. Can be empty for digital goods.
	FromAddress *Address
	// Customer's shipping/billing address.
	ToAddress Address
	// List of items in the transaction.
	TaxItems []TaxItem `json:"items"`
}

type TaxItem struct {
	// Amount of the transaction, in the smallest currency unit.
	// https://docs.stripe.com/currencies#minor-units
	Price decimal.Decimal
	// Quantity of the item.
	Quantity int
	// Tax Item Identifier
	Reference string `json:"reference"`
	// Tax code for the product.
	TaxCode string `json:"tax_code"`
}

type Address struct {
	Line1      string `json:"line1"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type CalculateTaxResponse struct {
	// The total amount of tax to be collected, in the smallest currency unit.
	Tax decimal.Decimal `json:"tax"`
	// The total amount after tax & shipping, in the smallest currency unit.
	// TotalAmount = Price + Tax + Shipping Rate
	TotalAmount decimal.Decimal `json:"total_amount"`
	// The currency of the amount_after_tax.
	Currency string `json:"currency"`
	// Breakdown of the tax calculation encoded in JSON.
	Breakdown json.JSON
}

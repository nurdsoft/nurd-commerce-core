package entities

import "github.com/shopspring/decimal"

type Dimensions struct {
	Length decimal.Decimal `json:"length,omitempty"`
	Width  decimal.Decimal `json:"width,omitempty"`
	Height decimal.Decimal `json:"height,omitempty"`
	Weight decimal.Decimal `json:"weight,omitempty"`
}

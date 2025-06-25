package entities

import (
	"time"

	"github.com/nurdsoft/nurd-commerce-core/shared/nullable"
)

type EstimateRatesResponse struct {
	RateType                  string                    `json:"rate_type,omitempty"`
	CarrierID                 string                    `json:"carrier_id,omitempty"`
	ShippingAmount            ShippingAmount            `json:"shipping_amount,omitempty"`
	InsuranceAmount           InsuranceAmount           `json:"insurance_amount,omitempty"`
	ConfirmationAmount        ConfirmationAmount        `json:"confirmation_amount,omitempty"`
	OtherAmount               OtherAmount               `json:"other_amount,omitempty"`
	RequestedComparisonAmount RequestedComparisonAmount `json:"requested_comparison_amount,omitempty"`
	RateDetails               []RateDetails             `json:"rate_details,omitempty"`
	Zone                      any                       `json:"zone,omitempty"`
	PackageType               any                       `json:"package_type,omitempty"`
	DeliveryDays              int                       `json:"delivery_days,omitempty"`
	GuaranteedService         bool                      `json:"guaranteed_service,omitempty"`
	// ShipEngine is returning null for this field
	EstimatedDeliveryDate nullable.NullTime `json:"estimated_delivery_date"`
	CarrierDeliveryDays   string            `json:"carrier_delivery_days,omitempty"`
	ShipDate              time.Time         `json:"ship_date,omitempty"`
	NegotiatedRate        bool              `json:"negotiated_rate,omitempty"`
	ServiceType           string            `json:"service_type,omitempty"`
	ServiceCode           string            `json:"service_code,omitempty"`
	Trackable             bool              `json:"trackable,omitempty"`
	CarrierCode           string            `json:"carrier_code,omitempty"`
	CarrierNickname       string            `json:"carrier_nickname,omitempty"`
	CarrierFriendlyName   string            `json:"carrier_friendly_name,omitempty"`
	ValidationStatus      string            `json:"validation_status,omitempty"`
	WarningMessages       []any             `json:"warning_messages,omitempty"`
	ErrorMessages         []string          `json:"error_messages,omitempty"`
}
type ShippingAmount struct {
	Currency string  `json:"currency,omitempty"`
	Amount   float64 `json:"amount,omitempty"`
}
type InsuranceAmount struct {
	Currency string  `json:"currency,omitempty"`
	Amount   float64 `json:"amount,omitempty"`
}
type ConfirmationAmount struct {
	Currency string  `json:"currency,omitempty"`
	Amount   float64 `json:"amount,omitempty"`
}
type OtherAmount struct {
	Currency string  `json:"currency,omitempty"`
	Amount   float64 `json:"amount,omitempty"`
}
type RequestedComparisonAmount struct {
	Currency string  `json:"currency,omitempty"`
	Amount   float64 `json:"amount,omitempty"`
}
type Amount struct {
	Currency string  `json:"currency,omitempty"`
	Amount   float64 `json:"amount,omitempty"`
}
type RateDetails struct {
	RateDetailType     string `json:"rate_detail_type,omitempty"`
	CarrierDescription string `json:"carrier_description,omitempty"`
	CarrierBillingCode string `json:"carrier_billing_code,omitempty"`
	CarrierMemo        any    `json:"carrier_memo,omitempty"`
	Amount             Amount `json:"amount,omitempty"`
	BillingSource      string `json:"billing_source,omitempty"`
}

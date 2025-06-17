package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/shared/json"
	"github.com/shopspring/decimal"
)

type OrderStatus string

func (o OrderStatus) String() string {
	return string(o)
}

const (
	Pending           OrderStatus = "pending"
	PaymentSuccess    OrderStatus = "payment_success"
	PaymentFailed     OrderStatus = "payment_failed"
	Processing        OrderStatus = "processing"
	Packed            OrderStatus = "packed"
	Shipped           OrderStatus = "shipped"
	FulfillmentFailed OrderStatus = "fulfillment_failed"
	Delivered         OrderStatus = "delivered"
	Cancelled         OrderStatus = "cancelled"
	ReturnRequested   OrderStatus = "return_requested"
	Returned          OrderStatus = "returned"
	Refunded          OrderStatus = "refunded"
)

type Order struct {
	ID                            uuid.UUID           `json:"id" gorm:"column:id"`
	CustomerID                    uuid.UUID           `json:"customer_id" gorm:"column:customer_id"`
	CartID                        uuid.UUID           `json:"cart_id" gorm:"column:cart_id"`
	OrderReference                string              `json:"order_reference" gorm:"column:order_reference"`
	TaxAmount                     decimal.Decimal     `json:"tax_amount" gorm:"column:tax_amount"`
	Subtotal                      decimal.Decimal     `json:"subtotal" gorm:"column:subtotal"`
	Total                         decimal.Decimal     `json:"total" gorm:"column:total"`
	Currency                      string              `json:"currency" gorm:"column:currency"`
	TaxBreakdown                  json.JSON           `json:"-" gorm:"column:tax_breakdown"`
	ShippingRate                  decimal.Decimal     `json:"shipping_rate" gorm:"column:shipping_rate"`
	ShippingCarrierName           string              `json:"shipping_carrier_name" gorm:"column:shipping_carrier_name"`
	ShippingCarrierCode           string              `json:"shipping_carrier_code" gorm:"column:shipping_carrier_code"`
	ShippingEstimatedDeliveryDate time.Time           `json:"shipping_estimated_delivery_date" gorm:"column:shipping_estimated_delivery_date"`
	ShippingBusinessDaysInTransit string              `json:"shipping_business_days_in_transit" gorm:"column:shipping_business_days_in_transit"`
	ShippingServiceType           string              `json:"shipping_service_type" gorm:"column:shipping_service_type"`
	ShippingServiceCode           string              `json:"shipping_service_code" gorm:"column:shipping_service_code"`
	DeliveryFullName              string              `json:"delivery_full_name" gorm:"column:delivery_full_name"`
	DeliveryAddress               string              `json:"delivery_address" gorm:"column:delivery_address"`
	DeliveryApartment             string              `json:"delivery_apartment" gorm:"column:delivery_apartment"`
	DeliveryCity                  *string             `json:"delivery_city" gorm:"column:delivery_city"`
	DeliveryStateCode             string              `json:"delivery_state_code" gorm:"column:delivery_state_code"`
	DeliveryCountryCode           string              `json:"delivery_country_code" gorm:"column:delivery_country_code"`
	DeliveryPostalCode            string              `json:"delivery_postal_code" gorm:"column:delivery_postal_code"`
	DeliveryPhoneNumber           *string             `json:"delivery_phone_number" gorm:"column:delivery_phone_number"`
	Status                        OrderStatus         `json:"status" db:"status"`
	FulfillmentMessage            *string             `json:"-" gorm:"column:fulfillment_message"`
	FulfillmentShipmentDate       *time.Time          `json:"-" gorm:"column:fulfillment_shipment_date"`
	FulfillmentFreightCharge      *decimal.Decimal    `json:"-" gorm:"column:fulfillment_freight_charge"`
	FulfillmentOrderTotal         *decimal.Decimal    `json:"-" gorm:"column:fulfillment_order_total"`
	FulfillmentAmountDue          *decimal.Decimal    `json:"-" gorm:"column:fulfillment_amount_due"`
	FulfillmentMetadata           json.JSON           `json:"-" gorm:"column:fulfillment_metadata"`
	FulfillmentTrackingNumber     *string             `json:"tracking_number,omitempty" gorm:"column:fulfillment_tracking_number"`
	SalesforceID                  string              `json:"-" gorm:"column:salesforce_id"`
	StripePaymentIntentID         *string             `json:"-" gorm:"column:stripe_payment_intent_id"`
	StripePaymentMethodID         string              `json:"stripe_payment_method_id" gorm:"column:stripe_payment_method_id"`
	AuthorizeNetPaymentID         *string             `json:"-" gorm:"column:authorizenet_payment_id"`
	CreatedAt                     time.Time           `json:"created_at" gorm:"column:created_at"`
	UpdatedAt                     time.Time           `json:"updated_at" gorm:"column:updated_at"`
	ItemsSummary                  []*OrderItemSummary `json:"items_summary,omitempty" gorm:"-"`
}

func (m *Order) TableName() string {
	return "orders"
}

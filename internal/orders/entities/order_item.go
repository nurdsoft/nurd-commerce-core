package entities

import (
	"time"

	"github.com/nurdsoft/nurd-commerce-core/shared/json"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderItemStatus string

func (o OrderItemStatus) String() string {
	return string(o)
}

const (
	ItemPending           OrderItemStatus = "pending"
	ItemProcessing        OrderItemStatus = "processing"
	ItemShipped           OrderItemStatus = "shipped"
	ItemFulfillmentFailed OrderItemStatus = "fulfillment_failed"
	ItemDelivered         OrderItemStatus = "delivered"
	ItemCancelled         OrderItemStatus = "cancelled"
	ItemReturnRequested   OrderItemStatus = "return_requested"
	ItemReturned          OrderItemStatus = "returned"
	ItemRefunded          OrderItemStatus = "refunded"
	ItemInitiatedRefund   OrderItemStatus = "initiated_refund"
)

// OrderItem represents an item in an order
type OrderItem struct {
	ID               uuid.UUID        `json:"id" gorm:"column:id;default:gen_random_uuid()"`
	OrderID          uuid.UUID        `json:"order_id" gorm:"column:order_id"`
	ProductID        uuid.UUID        `json:"product_id" gorm:"column:product_id"`
	ProductVariantID uuid.UUID        `json:"product_variant_id" gorm:"column:product_variant_id"`
	SKU              string           `json:"sku" gorm:"column:sku"`
	Description      *string          `json:"description" gorm:"column:description"`
	ImageURL         string           `json:"image_url" gorm:"column:image_url"`
	Name             string           `json:"name" gorm:"column:name"`
	Length           *decimal.Decimal `json:"length" gorm:"column:length"`
	Width            *decimal.Decimal `json:"width" gorm:"column:width"`
	Height           *decimal.Decimal `json:"height" gorm:"column:height"`
	Weight           *decimal.Decimal `json:"weight" gorm:"column:weight"`
	Quantity         int              `json:"quantity" gorm:"column:quantity"`
	Price            decimal.Decimal  `json:"price" gorm:"column:price"`
	Attributes       *json.JSON       `json:"attributes" db:"attributes"`
	// Shipping/Fulfillment fields
	ShippingRateID        *uuid.UUID       `json:"shipping_rate_id" gorm:"column:shipping_rate_id"`
	ShippingRate          *decimal.Decimal `json:"shipping_rate" gorm:"column:shipping_rate"`
	ShippingCarrierName   *string          `json:"shipping_carrier_name" gorm:"column:shipping_carrier_name"`
	ShippingCarrierCode   *string          `json:"shipping_carrier_code" gorm:"column:shipping_carrier_code"`
	ShippingServiceType   *string          `json:"shipping_service_type" gorm:"column:shipping_service_type"`
	ShippingServiceCode   *string          `json:"shipping_service_code" gorm:"column:shipping_service_code"`
	EstimatedDeliveryDate *time.Time       `json:"estimated_delivery_date" gorm:"column:estimated_delivery_date"`
	BusinessDaysInTransit *string          `json:"business_days_in_transit" gorm:"column:business_days_in_transit"`
	TrackingNumber        *string          `json:"tracking_number" gorm:"column:tracking_number"`
	TrackingURL           *string          `json:"tracking_url" gorm:"column:tracking_url"`
	ShipmentDate          *time.Time       `json:"shipment_date" gorm:"column:shipment_date"`
	FreightCharge         *decimal.Decimal `json:"freight_charge" gorm:"column:freight_charge"`
	AmountDue             *decimal.Decimal `json:"amount_due" gorm:"column:amount_due"`
	FulfillmentMessage    *string          `json:"fulfillment_message" gorm:"column:fulfillment_message"`
	FulfillmentMetadata   *json.JSON       `json:"fulfillment_metadata" gorm:"column:fulfillment_metadata"`
	CreatedAt             time.Time        `json:"created_at" gorm:"column:created_at"`
	UpdatedAt             time.Time        `json:"updated_at" gorm:"column:updated_at"`
	SalesforceID          string           `json:"-" gorm:"column:salesforce_id"`
	Status                OrderItemStatus  `json:"status" db:"status"`
	StripeRefundID        string           `json:"-" gorm:"column:stripe_refund_id"`
}

func (m *OrderItem) TableName() string {
	return "order_items"
}

// OrderItemSummary represents a summary of an order item (less detailed than OrderItem)
type OrderItemSummary struct {
	ID               uuid.UUID       `json:"id" gorm:"column:id;default:gen_random_uuid()"`
	ProductID        uuid.UUID       `json:"product_id" gorm:"column:product_id"`
	ProductVariantID uuid.UUID       `json:"product_variant_id" gorm:"column:product_variant_id"`
	SKU              string          `json:"sku" gorm:"column:sku"`
	ImageURL         string          `json:"image_url" gorm:"column:image_url"`
	Name             string          `json:"name" gorm:"column:name"`
	Quantity         int             `json:"quantity" gorm:"column:quantity"`
	Price            decimal.Decimal `json:"price" gorm:"column:price"`
	Attributes       *json.JSON      `json:"attributes" db:"attributes"`
	Status           OrderItemStatus `json:"status" db:"status"`
}

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
	CreatedAt        time.Time        `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time        `json:"updated_at" gorm:"column:updated_at"`
	SalesforceID     string           `json:"-" gorm:"column:salesforce_id"`
	Status           OrderItemStatus  `json:"status" db:"status"`
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

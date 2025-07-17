package entities

import (
	addressEntities "github.com/nurdsoft/nurd-commerce-core/internal/address/entities"
	cartEntities "github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	customerEntities "github.com/nurdsoft/nurd-commerce-core/internal/customer/entities"
	orderEntities "github.com/nurdsoft/nurd-commerce-core/internal/orders/entities"
)

type CreateInventoryOrderRequest struct {
	Order      orderEntities.Order
	OrderItems []*orderEntities.OrderItem
	Address    addressEntities.Address
	Customer   customerEntities.Customer
	CartItems  []cartEntities.CartItemDetail
}

type UpdateInventoryOrderStatusRequest struct {
	Order    orderEntities.Order
	Customer customerEntities.Customer
	Status   string
}

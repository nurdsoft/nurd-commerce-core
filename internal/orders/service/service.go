package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	cartEntities "github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	"math/rand/v2"
	"strings"
	"time"

	cartEntities "github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"

	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/internal/address/addressclient"
	addressEntities "github.com/nurdsoft/nurd-commerce-core/internal/address/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/cartclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/customer/customerclient"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/internal/orders/errors"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/repository"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/productclient"
	webhook "github.com/nurdsoft/nurd-commerce-core/internal/webhook/client"
	webhookEntities "github.com/nurdsoft/nurd-commerce-core/internal/webhook/entities"
	wishlistentities "github.com/nurdsoft/nurd-commerce-core/internal/wishlist/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/wishlistclient"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory"
	inventoryEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment"
	authorizenetEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/providers"
	stripeEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/entities"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type Service interface {
	CreateOrder(ctx context.Context, req *entities.CreateOrderRequest) (*entities.CreateOrderResponse, error)
	ListOrders(ctx context.Context, req *entities.ListOrdersRequest) (*entities.ListOrdersResponse, error)
	GetOrder(ctx context.Context, req *entities.GetOrderRequest) (*entities.GetOrderData, error)
	CancelOrder(ctx context.Context, req *entities.CancelOrderRequest) error
	ProcessPaymentSucceeded(ctx context.Context, paymentID string) error
	ProcessPaymentFailed(ctx context.Context, paymentID string) error
	UpdateOrder(ctx context.Context, req *entities.UpdateOrderRequest) error
	RefundOrder(ctx context.Context, req *entities.RefundOrderRequest) (*entities.RefundOrderResponse, error)
	ProcessRefundSucceeded(ctx context.Context, refundId string, refundAmount decimal.Decimal) error
}

type service struct {
	repo            repository.Repository
	log             *zap.SugaredLogger
	customerClient  customerclient.Client
	cartClient      cartclient.Client
	paymentClient   payment.Client
	wishlistClient  wishlistclient.Client
	inventoryClient inventory.Client
	addressClient   addressclient.Client
	productClient   productclient.Client
	webhookClient   webhook.Client
	config          cfg.Config
}

func New(
	repo repository.Repository, log *zap.SugaredLogger, customerClient customerclient.Client,
	cartClient cartclient.Client, paymentClient payment.Client,
	wishlistClient wishlistclient.Client, config cfg.Config,
	inventoryClient inventory.Client, addressClient addressclient.Client, productClient productclient.Client,
	webhookClient webhook.Client,
) Service {
	return &service{
		repo:            repo,
		log:             log,
		customerClient:  customerClient,
		cartClient:      cartClient,
		paymentClient:   paymentClient,
		wishlistClient:  wishlistClient,
		inventoryClient: inventoryClient,
		addressClient:   addressClient,
		productClient:   productClient,
		webhookClient:   webhookClient,
		config:          config,
	}
}

// swagger:route POST /orders orders CreateOrderRequest
//
// # Create Order
// ### Create an order based on active cart items
//
// Produces:
// - application/json
//
// Responses:
//
//	200: CreateOrderResponse Order created successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) CreateOrder(ctx context.Context, req *entities.CreateOrderRequest) (*entities.CreateOrderResponse, error) {
	customerID, err := uuid.Parse(sharedMeta.XCustomerID(ctx))
	if err != nil {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	address, err := s.addressClient.GetAddress(ctx, &addressEntities.GetAddressRequest{
		AddressID: req.Body.AddressID,
	})
	if err != nil {
		return nil, err
	}

	// get user's active cart via cartclient
	cart, err := s.cartClient.GetCart(ctx)
	if err != nil {
		return nil, err
	}

	// get cart items via cartclient
	cartItems, err := s.cartClient.GetCartItems(ctx)
	if err != nil {
		return nil, err
	}

	// get shipping rate via cartclient
	var shipping *cartEntities.CartShippingRate = nil
	if req.Body.ShippingRateID != nil {
		shipping, err = s.cartClient.GetShippingRateByID(ctx, *req.Body.ShippingRateID)
		if err != nil {
			return nil, err
		}
		if cart.ShippingRateID.String() != shipping.Id.String() {
			s.log.Errorln("shipping rate does not match cart shipping rate")
			return nil, moduleErrors.NewAPIError("ORDER_ERROR_CREATING")
		}

	}

	orderItems := []*entities.OrderItem{}
	orderId := uuid.New()

	var subTotal decimal.Decimal

	for _, item := range cartItems.Items {
		orderItem := &entities.OrderItem{
			ID:               uuid.New(),
			OrderID:          orderId,
			ProductID:        item.ProductID,
			ProductVariantID: item.ProductVariantID,
			SKU:              item.SKU,
			Description:      item.Description,
			ImageURL:         item.ImageURL,
			Name:             item.Name,
			Height:           item.Height,
			Length:           item.Length,
			Width:            item.Width,
			Weight:           item.Weight,
			Quantity:         item.Quantity,
			Price:            item.Price,
			Attributes:       item.Attributes,
			Status:           entities.ItemPending,
		}
		orderItems = append(orderItems, orderItem)
		subTotal = subTotal.Add(item.Price.Mul(decimal.NewFromInt(int64(item.Quantity))))
	}
	total := cart.TaxAmount.Add(subTotal)
	if shipping != nil {
		total = total.Add(shipping.Amount)
	}

	customer, err := s.customerClient.GetCustomer(ctx)
	if err != nil {
		return nil, err
	}

	paymentReq := entities.CreatePaymentRequest{
		Amount:          total,
		Currency:        cart.TaxCurrency,
		Customer:        *customer,
		PaymentMethodId: req.Body.StripePaymentMethodID,
		PaymentNonce:    req.Body.PaymentNonce,
		BillingInfo:     req.Body.BillingInfo,
	}

	paymentResponse, err := s.createPaymentByProvider(ctx, paymentReq)
	if err != nil {
		return nil, err
	}

	orderStatus := entities.Pending
	switch paymentResponse.Status {
	case providers.PaymentStatusSuccess:
		orderStatus = entities.PaymentSuccess
	case providers.PaymentStatusFailed:
		orderStatus = entities.PaymentFailed
	}

	orderRef, err := s.generateOrderRef(ctx, orderId.String())
	if err != nil {
		return nil, err
	}

	order := &entities.Order{
		ID:                  orderId,
		CustomerID:          customerID,
		CartID:              cart.Id,
		OrderReference:      orderRef,
		TaxAmount:           cart.TaxAmount,
		Subtotal:            subTotal,
		Total:               total,
		Currency:            cart.TaxCurrency,
		TaxBreakdown:        cart.TaxBreakdown,
		DeliveryFullName:    address.FullName,
		DeliveryAddress:     address.Address,
		DeliveryCity:        address.City,
		DeliveryStateCode:   address.StateCode,
		DeliveryCountryCode: address.CountryCode,
		DeliveryPostalCode:  address.PostalCode,
		DeliveryPhoneNumber: address.PhoneNumber,
		Status:              orderStatus,
	}

	if shipping != nil {
		order.ShippingRate = &shipping.Amount
		order.ShippingCarrierName = &shipping.CarrierName
		order.ShippingCarrierCode = &shipping.CarrierCode
		order.ShippingEstimatedDeliveryDate = &shipping.EstimatedDeliveryDate
		order.ShippingBusinessDaysInTransit = &shipping.BusinessDaysInTransit
		order.ShippingServiceType = &shipping.ServiceType
		order.ShippingServiceCode = &shipping.ServiceCode
	}

	switch s.paymentClient.GetProvider() {
	case providers.ProviderStripe:
		order.StripePaymentIntentID = &paymentResponse.ID
		order.StripePaymentMethodID = req.Body.StripePaymentMethodID
	case providers.ProviderAuthorizeNet:
		order.AuthorizeNetPaymentID = &paymentResponse.ID
	}

	// create order
	err = s.repo.CreateOrder(ctx, cart.Id, order, orderItems)
	if err != nil {
		return nil, moduleErrors.NewAPIError("ORDER_ERROR_CREATING")
	} else {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			err := s.webhookClient.NotifyOrderStatusChange(bgCtx, &webhookEntities.NotifyOrderStatusChangeRequest{
				CustomerID:     customerID.String(),
				OrderID:        order.ID.String(),
				OrderReference: order.OrderReference,
				Status:         orderStatus.String(),
			})
			if err != nil {
				s.log.Errorf("Error notifying order status change: %v", err)
			}
		}()

		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			_, err := s.inventoryClient.CreateOrder(bgCtx, inventoryEntities.CreateInventoryOrderRequest{
				Order:      *order,
				OrderItems: orderItems,
				Address:    *address,
				Customer:   *customer,
				CartItems:  cartItems.Items,
			})
			if err != nil {
				s.log.Errorf("Error creating order on inventory: %v", err)
			}
		}()
	}

	return &entities.CreateOrderResponse{
		OrderReference: order.OrderReference,
	}, nil
}

func (s *service) createPaymentByProvider(ctx context.Context, paymentReq entities.CreatePaymentRequest) (providers.PaymentProviderResponse, error) {
	var req any

	switch s.paymentClient.GetProvider() {
	case providers.ProviderStripe:
		req = stripeEntities.CreatePaymentIntentRequest{
			Amount:          paymentReq.Amount,
			Currency:        paymentReq.Currency,
			CustomerId:      paymentReq.Customer.StripeID,
			PaymentMethodId: paymentReq.PaymentMethodId,
		}
	case providers.ProviderAuthorizeNet:
		req = authorizenetEntities.CreatePaymentTransactionRequest{
			Amount:       paymentReq.Amount,
			PaymentNonce: paymentReq.PaymentNonce,
			BillingInfo: authorizenetEntities.BillingInfo{
				FirstName: paymentReq.BillingInfo.FirstName,
				LastName:  paymentReq.BillingInfo.LastName,
				Address:   paymentReq.BillingInfo.Address,
				City:      paymentReq.BillingInfo.City,
				State:     paymentReq.BillingInfo.State,
				Country:   paymentReq.BillingInfo.Country,
				Zip:       paymentReq.BillingInfo.Zip,
			},
		}
	}

	res, err := s.paymentClient.CreatePayment(ctx, req)
	if err != nil {
		return providers.PaymentProviderResponse{}, err
	}

	return res, nil
}

// swagger:route GET /orders orders ListOrdersRequest
//
// # List Orders
// ### List all of customer's orders
//
// Produces:
// - application/json
//
// Responses:
//
//	200: ListOrdersResponse Orders listed successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) ListOrders(ctx context.Context, req *entities.ListOrdersRequest) (*entities.ListOrdersResponse, error) {
	customerID, err := uuid.Parse(sharedMeta.XCustomerID(ctx))
	if err != nil {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	orders, nextCursor, err := s.repo.ListOrders(
		ctx,
		customerID,
		req.Limit,
		req.Cursor,
		req.IncludeItems,
	)

	if err != nil {
		return nil, moduleErrors.NewAPIError("ORDER_ERROR_LISTING")
	}

	return &entities.ListOrdersResponse{
		Orders:     orders,
		NextCursor: nextCursor,
	}, nil
}

// swagger:route GET /orders/{order_id} orders GetOrderRequest
//
// # Get Order with items
//
// Produces:
// - application/json
//
// Responses:
//
//	200: GetOrderResponse Order fetched successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) GetOrder(ctx context.Context, req *entities.GetOrderRequest) (*entities.GetOrderData, error) {
	customerID, err := uuid.Parse(sharedMeta.XCustomerID(ctx))
	if err != nil {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	orderId, err := uuid.Parse(req.OrderID.String())
	if err != nil {
		return nil, moduleErrors.NewAPIError("ORDER_ID_REQUIRED")
	}

	order, err := s.repo.GetOrderByID(ctx, orderId)
	if err != nil {
		return nil, moduleErrors.NewAPIError("ORDER_ERROR_GETTING")
	}

	if order.CustomerID != customerID {
		return nil, moduleErrors.NewAPIError("ORDER_NOT_FOUND")
	}

	orderItems, err := s.repo.GetOrderItemsByID(ctx, orderId)
	if err != nil {
		return nil, moduleErrors.NewAPIError("ORDER_ERROR_GETTING_ITEMS")
	}

	return &entities.GetOrderData{
		Order:      order,
		OrderItems: orderItems,
	}, nil
}

// swagger:route DELETE /orders/{order_id} orders CancelOrderRequest
//
// # Cancel Order
// ### Cancel an order
//
// Produces:
// - application/json
//
// Responses:
//
//	200: DefaultResponse Order canceled successfully
//	304: DefaultError Order is already cancelled
//	400: DefaultError Bad Request
//	404: DefaultError Order not found
//	500: DefaultError Internal Server Error
func (s *service) CancelOrder(ctx context.Context, req *entities.CancelOrderRequest) error {
	customerID, err := uuid.Parse(sharedMeta.XCustomerID(ctx))
	if err != nil {
		return moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	orderId, err := uuid.Parse(req.OrderID.String())
	if err != nil {
		return moduleErrors.NewAPIError("ORDER_ID_REQUIRED")
	}

	order, err := s.repo.GetOrderByID(ctx, orderId)
	if err != nil {
		return moduleErrors.NewAPIError("ORDER_ERROR_GETTING")
	}

	if order.CustomerID != customerID {
		return moduleErrors.NewAPIError("ORDER_NOT_FOUND")
	}

	switch order.Status {
	case entities.Pending:
		// cancel order with no refund required
		err = s.repo.Update(ctx, map[string]interface{}{
			"status": entities.Cancelled,
		}, order.ID.String(), order.CustomerID.String())
		if err != nil {
			return moduleErrors.NewAPIError("ORDER_ERROR_CANCELLING")
		}
	case entities.PaymentSuccess:
		err = s.repo.Update(ctx, map[string]interface{}{
			"status": entities.Cancelled,
		}, order.ID.String(), order.CustomerID.String())
		if err != nil {
			return moduleErrors.NewAPIError("ORDER_ERROR_CANCELLING")
		}
		// TODO refund payment via stripe & send email to user
	case entities.Cancelled:
		return moduleErrors.NewAPIError("ORDER_IS_ALREADY_CANCELLED")
	default:
		return moduleErrors.NewAPIError("ORDER_CANNOT_BE_CANCELLED")
	}

	customer, err := s.customerClient.GetCustomerByID(ctx, order.CustomerID.String())
	if err != nil {
		return err
	}

	// update order status on inventory
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		err = s.inventoryClient.UpdateOrderStatus(bgCtx, inventoryEntities.UpdateInventoryOrderStatusRequest{
			Order:    *order,
			Customer: *customer,
			Status:   entities.Cancelled.String(),
		})
		if err != nil {
			s.log.Errorf("Error updating order status on inventory: %v", err)
		}
	}()

	return nil
}

func (s *service) ProcessPaymentSucceeded(ctx context.Context, paymentID string) error {
	order, err := s.getOrderByPaymentID(ctx, paymentID)
	if err != nil {
		return moduleErrors.NewAPIError("ORDER_NOT_FOUND_BY_PAYMENT_ID")
	}

	if order.Status != entities.Pending {
		return moduleErrors.NewAPIError("ORDER_IS_NOT_PENDING")
	}

	err = s.repo.Update(ctx, map[string]interface{}{
		"status": entities.PaymentSuccess,
	}, order.ID.String(), order.CustomerID.String())
	if err != nil {
		return err
	}

	customer, err := s.customerClient.GetCustomerByID(ctx, order.CustomerID.String())
	if err != nil {
		return err
	}

	items, err := s.repo.GetOrderItemsByID(ctx, order.ID)
	if err != nil {
		return err
	}

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		err := s.webhookClient.NotifyOrderStatusChange(bgCtx, &webhookEntities.NotifyOrderStatusChangeRequest{
			CustomerID:     order.CustomerID.String(),
			OrderID:        order.ID.String(),
			OrderReference: order.OrderReference,
			Status:         entities.PaymentSuccess.String(),
		})
		if err != nil {
			s.log.Errorf("Error notifying order status change: %v", err)
		}
	}()

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		err := s.inventoryClient.UpdateOrderStatus(bgCtx, inventoryEntities.UpdateInventoryOrderStatusRequest{
			Order:    *order,
			Customer: *customer,
			Status:   entities.PaymentSuccess.String(),
		})
		if err != nil {
			s.log.Errorf("Error updating order status on inventory: %v", err)
		}
	}()

	var productIDs []uuid.UUID

	for _, item := range items {
		productVariant, err := s.productClient.GetProductVariantByID(ctx, item.ProductVariantID.String())
		if err != nil {
			s.log.Errorf("Error fetching product variant: %v", err)
			continue
		}
		productIDs = append(productIDs, productVariant.ProductID)
	}

	// once a user has purchased a product, remove it from their wishlist
	err = s.wishlistClient.BulkRemoveFromWishlist(ctx, &wishlistentities.BulkRemoveFromWishlistRequest{
		CustomerID: customer.ID,
		ProductIDs: productIDs,
	})
	if err != nil {
		// non-destructive error, the process should continue in case of failure
		s.log.Errorf("Error removing products from wishlist: %v", err)
	}

	return nil
}

func (s *service) ProcessPaymentFailed(ctx context.Context, paymentID string) error {
	order, err := s.getOrderByPaymentID(ctx, paymentID)
	if err != nil {
		return moduleErrors.NewAPIError("ORDER_NOT_FOUND_BY_PAYMENT_ID")
	}

	if order.Status != entities.Pending {
		return moduleErrors.NewAPIError("ORDER_IS_NOT_PENDING")
	}

	err = s.repo.Update(ctx, map[string]interface{}{
		"status": entities.PaymentFailed,
	}, order.ID.String(), order.CustomerID.String())

	if err != nil {
		s.log.Errorf("Error updating order status: %v", err)
		return err
	}

	customer, err := s.customerClient.GetCustomerByID(ctx, order.CustomerID.String())
	if err != nil {
		s.log.Errorf("Error fetching customer: %v", err)
		return err
	}

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		err := s.webhookClient.NotifyOrderStatusChange(bgCtx, &webhookEntities.NotifyOrderStatusChangeRequest{
			CustomerID:     order.CustomerID.String(),
			OrderID:        order.ID.String(),
			OrderReference: order.OrderReference,
			Status:         entities.PaymentFailed.String(),
		})
		if err != nil {
			s.log.Errorf("Error notifying order status change: %v", err)
		}
	}()

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		err := s.inventoryClient.UpdateOrderStatus(bgCtx, inventoryEntities.UpdateInventoryOrderStatusRequest{
			Order:    *order,
			Customer: *customer,
			Status:   entities.PaymentFailed.String(),
		})
		if err != nil {
			s.log.Errorf("Error updating order status on inventory: %v", err)
		}
	}()

	return nil
}

func (s *service) getOrderByPaymentID(ctx context.Context, paymentID string) (*entities.Order, error) {
	switch s.paymentClient.GetProvider() {
	case providers.ProviderStripe:
		return s.repo.GetOrderByStripePaymentIntentID(ctx, paymentID)
	case providers.ProviderAuthorizeNet:
		return s.repo.GetOrderByAuthorizeNetPaymentID(ctx, paymentID)
	default:
		// this should never happen, but just in case
		return nil, errors.New("payment provider not supported")
	}
}

// swagger:route PUT /orders/{order_reference} orders UpdateOrderRequest
//
// # Update Order
// ### Update an order
//
// Produces:
// - application/json
//
// Responses:
//
//	200: Order updated successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) UpdateOrder(ctx context.Context, req *entities.UpdateOrderRequest) error {
	order, err := s.repo.GetOrderByReference(ctx, req.OrderReference)
	if err != nil {
		s.log.Errorf("Error fetching order: %v", err)
		return err
	}

	// update order status
	data := map[string]interface{}{}

	if req.Body.Status != nil {
		data["status"] = req.Body.Status
	}

	if req.Body.FulfillmentShipmentDate != nil {
		data["fulfillment_shipment_date"] = req.Body.FulfillmentShipmentDate
	}

	if req.Body.FulfillmentFreightCharge != nil {
		data["fulfillment_freight_charge"] = req.Body.FulfillmentFreightCharge
	}

	if req.Body.FulfillmentOrderTotal != nil {
		data["fulfillment_order_total"] = req.Body.FulfillmentOrderTotal
	}

	if req.Body.FulfillmentMessage != nil {
		data["fulfillment_message"] = req.Body.FulfillmentMessage
	}

	if req.Body.FulfillmentAmountDue != nil {
		data["fulfillment_amount_due"] = req.Body.FulfillmentAmountDue
	}

	if req.Body.FulfilmentMetadata != nil {
		data["fulfillment_metadata"] = req.Body.FulfilmentMetadata
	}

	if req.Body.FulfillmentTrackingNumber != nil {
		data["fulfillment_tracking_number"] = req.Body.FulfillmentTrackingNumber
	}

	if req.Body.FulfillmentTrackingURL != nil {
		data["fulfillment_tracking_url"] = req.Body.FulfillmentTrackingURL
	}

	if len(req.Body.Items) > 0 {
		itemsData := make([]map[string]interface{}, 0, len(req.Body.Items))
		for _, item := range req.Body.Items {
			itemData := map[string]interface{}{}

			// Add ID if provided
			if item.ID != "" {
				itemData["id"] = item.ID
			}

			// Add SKU if provided (keep original key for identification)
			if item.Sku != "" {
				itemData["sku"] = item.Sku
			}

			// Add status if provided
			if item.Status != nil {
				itemData["status"] = item.Status
			}

			itemsData = append(itemsData, itemData)
		}
		data["items"] = itemsData
	}

	s.log.Infof("Updating order %s with data: %v", order.ID.String(), data)
	err = s.repo.Update(ctx, data, order.ID.String(), order.CustomerID.String())

	if err != nil {
		s.log.Errorf("Error updating order status: %v", err)
		return err
	}

	// Notify only when the status has changed
	if req.Body.Status != nil && order.Status.String() != *req.Body.Status {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()

			err := s.webhookClient.NotifyOrderStatusChange(bgCtx, &webhookEntities.NotifyOrderStatusChangeRequest{
				CustomerID:     order.CustomerID.String(),
				OrderID:        order.ID.String(),
				OrderReference: order.OrderReference,
				Status:         *req.Body.Status,
			})
			if err != nil {
				s.log.Errorf("Error notifying order status change: %v", err)
			}
		}()
	}

	return nil
}

// swagger:route POST /orders/{order_reference}/refund orders RefundOrderRequest
//
// # Initiate an Order Refund
// ### Refund order items by SKU & Quantity
//
// Produces:
// - application/json
//
// Responses:
//
//	200: RefundOrderResponse Order refund initiated successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) RefundOrder(ctx context.Context, req *entities.RefundOrderRequest) (*entities.RefundOrderResponse, error) {
	provider := s.paymentClient.GetProvider()

	order, err := s.repo.GetOrderByReference(ctx, req.OrderReference)
	if err != nil {
		return nil, moduleErrors.NewAPIError("ORDER_NOT_FOUND")
	}

	// TODO: Improve by adding a FSM for Order State Machine

	// disable multiple refunds for the same order
	if order.Status == entities.Refunded {
		return nil, moduleErrors.NewAPIError("ORDER_REFUNDING_ERROR", "Order is not eligible for refund")
	}

	// get order items
	orderItems, err := s.repo.GetOrderItemsByID(ctx, order.ID)
	if err != nil {
		s.log.Errorf("Error fetching order items: %v", err)
		return nil, moduleErrors.NewAPIError("ORDER_ERROR_GETTING_ITEMS")
	}

	// iterate over order items and request body items to check if they match by sku
	var refundableAmount decimal.Decimal
	var shouldChangeOrderStatus bool
	var shouldRefundWholeOrder bool
	var totalRefundableQuantity, totalOrderQuantity int
	orderItemsRefundData := make(map[string]interface{})
	orderRefundData := make(map[string]interface{})
	refundableItems := make([]*entities.RefundableItem, 0)
	refundableOrderItems := make([]*entities.OrderItem, 0)

	for _, orderItem := range orderItems {
		// Calculate total order quantity
		totalOrderQuantity += orderItem.Quantity

		// Extra work, filter order items that can be refunded
		if orderItem.Status != entities.ItemRefunded && orderItem.Status != entities.ItemInitiatedRefund {
			refundableOrderItems = append(refundableOrderItems, orderItem)
		}
	}

	for _, item := range req.Body.Items {
		if item.Sku != "" {
			for _, orderItem := range refundableOrderItems {
				if orderItem.SKU == item.Sku && orderItem.Quantity >= item.Quantity {
					totalItemCost := orderItem.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
					refundableAmount = refundableAmount.Add(totalItemCost)
					// quantity of items that are valid for refund
					totalRefundableQuantity += item.Quantity

					switch provider {
					case providers.ProviderStripe:
						orderItemsRefundData[orderItem.ID.String()] = map[string]interface{}{
							"status":               entities.ItemInitiatedRefund.String(),
							"stripe_refund_amount": totalItemCost.InexactFloat64(),
						}
					}
					refundableItems = append(refundableItems, &entities.RefundableItem{
						ItemId:   orderItem.ID.String(),
						Sku:      orderItem.SKU,
						Quantity: item.Quantity,
						Price:    orderItem.Price,
					})
					break
				}
			}
		}
	}

	if refundableAmount.IsZero() {
		return nil, moduleErrors.NewAPIError("ORDER_REFUNDING_ERROR", "No refundable items found or amount is zero")
	}

	if refundableAmount.GreaterThan(order.Total) {
		return nil, moduleErrors.NewAPIError("ORDER_REFUNDING_ERROR", "Refundable amount exceeds order total")
	}

	// Check if we're refunding all available order items
	if totalRefundableQuantity == totalOrderQuantity {
		shouldChangeOrderStatus = true
		shouldRefundWholeOrder = true
	}

	switch provider {
	case providers.ProviderStripe:
		var refundResponse *providers.RefundResponse
		var err error

		if shouldRefundWholeOrder {
			// Everything needs to be refunded, including shipping and taxes
			s.log.Info("Refunding entire order amount via Stripe")
			refundResponse, err = s.paymentClient.Refund(ctx, &stripeEntities.RefundRequest{
				PaymentIntentId: *order.StripePaymentIntentID,
			})
		} else {
			// TODO: find a way to calculate the item-level refunding amount excluding shipping and taxes
			s.log.Infof("Refunding partial order amount via Stripe: %s", refundableAmount.String())
			refundResponse, err = s.paymentClient.Refund(ctx, &stripeEntities.RefundRequest{
				PaymentIntentId: *order.StripePaymentIntentID,
				Amount:          refundableAmount,
			})
		}
		if err != nil {
			s.log.Errorf("Error processing refund via Stripe: %v", err)
			return nil, moduleErrors.NewAPIError("ORDER_REFUNDING_ERROR", "Error processing refund via Stripe")
		}

		if refundResponse.Status != stripeEntities.StripeRefundSucceeded && refundResponse.Status != stripeEntities.StripeRefundPending {
			s.log.Errorf("Stripe refund failed with status: %s and ID: %s", refundResponse.Status, refundResponse.ID)
			return nil, moduleErrors.NewAPIError("ORDER_REFUNDING_ERROR", "Stripe refund failed")
		}
		for orderItemID, data := range orderItemsRefundData {
			data.(map[string]interface{})["stripe_refund_id"] = refundResponse.ID
			data.(map[string]interface{})["stripe_refund_created_at"] = time.Now().UTC()
			orderItemsRefundData[orderItemID] = data
		}
		var stripeTotalRefund decimal.Decimal

		if shouldChangeOrderStatus {
			// update order refund total for the whole order
			if order.StripeRefundTotal != nil {
				stripeTotalRefund = order.StripeRefundTotal.Add(refundableAmount)
			} else {
				stripeTotalRefund = refundableAmount
			}
			// this total represents the total amount refunded via Stripe, it can change with each initiated refund
			orderRefundData["stripe_refund_total"] = stripeTotalRefund
			orderRefundData["status"] = entities.Refunded
		}

		// Stripe doesn't care about individual item refunds, it just processes the amount given
		// so its safe to mark all refundable items as refunded, assuming the above call succeeded
		for _, item := range refundableItems {
			if item.ItemId != "" {
				item.RefundInitiated = true
			}
		}
	default:
		return nil, moduleErrors.NewAPIError("ORDER_REFUNDING_ERROR", "Payment provider not supported for refunds")
	}

	// update order items with refund data
	err = s.repo.UpdateOrderWithOrderItems(ctx, order.ID, orderRefundData, orderItemsRefundData)
	if err != nil {
		s.log.Errorf("Error updating order items with refund data: %v", err)
	}

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		err := s.webhookClient.NotifyOrderStatusChange(bgCtx, &webhookEntities.NotifyOrderStatusChangeRequest{
			CustomerID:     order.CustomerID.String(),
			OrderID:        order.ID.String(),
			OrderReference: order.OrderReference,
			Status:         entities.Refunded.String(),
		})
		if err != nil {
			s.log.Errorf("Error notifying order status change: %v", err)
		}
	}()

	return &entities.RefundOrderResponse{
		TotalRefundableAmount: refundableAmount,
		RefundableItems:       refundableItems,
	}, nil
}

func (s *service) ProcessRefundSucceeded(ctx context.Context, refundId string, refundAmount decimal.Decimal) error {
	s.log.Info("Processing refund succeeded for refund ID: %s with amount: %s", refundId, refundAmount.String())

	orderItems, err := s.repo.GetOrderItemsByStripeRefundID(ctx, refundId)
	if err != nil {
		s.log.Errorf("Error fetching order items by refund ID: %v", err)
		return moduleErrors.NewAPIError("ORDER_ITEMS_NOT_FOUND_BY_REFUND_ID")
	}

	if len(orderItems) == 0 {
		s.log.Errorf("No order items found for refund ID: %s", refundId)
		return nil
	}

	// Assuming refundId is unique for each order item, all the resulting order items should have the same order ID
	orderID := orderItems[0].OrderID
	orderItemsRefundData := make(map[string]interface{})

	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		s.log.Errorf("Error fetching order by ID: %v", err)
		return moduleErrors.NewAPIError("ORDER_NOT_FOUND_BY_ID")
	}

	if order.Status == entities.Refunded {
		s.log.Errorf("Order %s is already in refunded state", orderID)
		return nil
	}

	var shouldChangeOrderStatus bool
	var itemsRefunded int
	orderRefundData := make(map[string]interface{})

	// gather count of items that are already refunded
	for _, item := range orderItems {
		if item.Status == entities.ItemRefunded {
			itemsRefunded++
		}
	}

	paymentProvider := s.paymentClient.GetProvider()

	switch paymentProvider {
	case providers.ProviderStripe:
		for _, item := range orderItems {
			if item.StripeRefundID == refundId && item.Status == entities.ItemInitiatedRefund {
				orderItemsRefundData[item.ID.String()] = map[string]interface{}{
					"status":               entities.ItemRefunded,
					"stripe_refund_amount": refundAmount.InexactFloat64(),
				}
				itemsRefunded++
			}
		}

		if itemsRefunded == len(orderItems) {
			shouldChangeOrderStatus = true
		}

		if shouldChangeOrderStatus {
			// refund status for partial refunds is only available at item level
			// the order status will be set to refunded only if all items are refunded)
			orderRefundData["status"] = entities.Refunded

			if order.StripeRefundTotal != nil {
				orderRefundData["stripe_refund_total"] = order.StripeRefundTotal.Add(refundAmount)
			} else {
				orderRefundData["stripe_refund_total"] = refundAmount
			}
		}
	default:
		s.log.Errorf("Payment provider %s does not support refund processing", paymentProvider)
		return nil
	}

	// update order items with refund data
	err = s.repo.UpdateOrderWithOrderItems(ctx, orderID, orderRefundData, orderItemsRefundData)
	if err != nil {
		s.log.Errorf("Error updating order items with refund data: %v", err)
	}

	return nil
}

// generateOrderRef generates a unique order reference based on the order ID.
func (s *service) generateOrderRef(ctx context.Context, orderId string) (string, error) {
	for {
		ref := generateAlphanumericOrderRef(orderId)

		// Check if it already exists
		exists, err := s.repo.OrderReferenceExists(ctx, ref)
		if err != nil {
			s.log.Errorf("Error checking order reference existence: %v", err)
			return "", moduleErrors.NewAPIError("ORDER_ERROR_CREATING")
		}

		if !exists {
			return ref, nil
		}

		// Append randomness to orderId to alter the hash in case of collision
		orderId += fmt.Sprintf("%d", rand.IntN(99999))
	}
}

// generateAlphanumericOrderRef generates a 10-character alphanumeric string based on the order ID.
func generateAlphanumericOrderRef(orderId string) string {
	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(orderId))

	// Define alphanumeric character set
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	var result strings.Builder
	result.Grow(10)

	// Use bytes from hash to select characters from charset
	for i := 0; i < 10; i++ {
		// Use modulo to map hash byte to character set index
		charIndex := hash[i] % byte(len(charset))
		result.WriteByte(charset[charIndex])
	}

	return result.String()
}

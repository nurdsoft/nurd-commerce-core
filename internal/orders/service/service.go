package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

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
	salesforce "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	salesforceEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/entities"
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
}

type service struct {
	repo             repository.Repository
	log              *zap.SugaredLogger
	customerClient   customerclient.Client
	cartClient       cartclient.Client
	paymentClient    payment.Client
	wishlistClient   wishlistclient.Client
	salesforceClient salesforce.Client
	addressClient    addressclient.Client
	productClient    productclient.Client
	webhookClient    webhook.Client
	config           cfg.Config
}

func New(
	repo repository.Repository, log *zap.SugaredLogger, customerClient customerclient.Client,
	cartClient cartclient.Client, paymentClient payment.Client,
	wishlistClient wishlistclient.Client, config cfg.Config,
	salesforceClient salesforce.Client, addressClient addressclient.Client, productClient productclient.Client,
	webhookClient webhook.Client,
) Service {
	return &service{
		repo:             repo,
		log:              log,
		customerClient:   customerClient,
		cartClient:       cartClient,
		paymentClient:    paymentClient,
		wishlistClient:   wishlistClient,
		salesforceClient: salesforceClient,
		addressClient:    addressClient,
		productClient:    productClient,
		webhookClient:    webhookClient,
		config:           config,
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
	shipping, err := s.cartClient.GetShippingRateByID(ctx, req.Body.ShippingRateID)
	if err != nil {
		return nil, err
	}

	if cart.ShippingRateID != req.Body.ShippingRateID && cart.ShippingRateID != shipping.Id {
		s.log.Errorln("shipping rate does not match cart shipping rate")
		return nil, moduleErrors.NewAPIError("ORDER_ERROR_CREATING")
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
		}
		orderItems = append(orderItems, orderItem)
		subTotal = subTotal.Add(item.Price.Mul(decimal.NewFromInt(int64(item.Quantity))))
	}

	total := cart.TaxAmount.Add(shipping.Amount).Add(subTotal)

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
		ID:                            orderId,
		CustomerID:                    customerID,
		CartID:                        cart.Id,
		OrderReference:                orderRef,
		TaxAmount:                     cart.TaxAmount,
		Subtotal:                      subTotal,
		Total:                         total,
		Currency:                      cart.TaxCurrency,
		TaxBreakdown:                  cart.TaxBreakdown,
		ShippingRate:                  shipping.Amount,
		ShippingCarrierName:           shipping.CarrierName,
		ShippingCarrierCode:           shipping.CarrierCode,
		ShippingEstimatedDeliveryDate: shipping.EstimatedDeliveryDate,
		ShippingBusinessDaysInTransit: shipping.BusinessDaysInTransit,
		ShippingServiceType:           shipping.ServiceType,
		ShippingServiceCode:           shipping.ServiceCode,
		DeliveryFullName:              address.FullName,
		DeliveryAddress:               address.Address,
		DeliveryCity:                  address.City,
		DeliveryStateCode:             address.StateCode,
		DeliveryCountryCode:           address.CountryCode,
		DeliveryPostalCode:            address.PostalCode,
		DeliveryPhoneNumber:           address.PhoneNumber,
		Status:                        orderStatus,
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
			err = s.webhookClient.NotifyOrderStatusChange(bgCtx, &webhookEntities.NotifyOrderStatusChangeRequest{
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
			if customer.SalesforceID == nil {
				s.log.Errorf("Customer does not have a salesforce id")
				return
			}

			bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()

			city := ""
			if address.City != nil {
				city = *address.City
			}

			// create order on salesforce
			res, err := s.salesforceClient.CreateOrder(bgCtx, &salesforceEntities.CreateSFOrderRequest{
				OrderReferenceC:         order.OrderReference,
				AccountID:               *customer.SalesforceID,
				EffectiveDate:           time.Now().Format("2006-01-02"),
				Status:                  entities.Pending.String(),
				BillingStreet:           address.Address,
				BillingCity:             city,
				BillingState:            address.StateCode,
				BillingPostalCode:       address.PostalCode,
				BillingCountry:          address.CountryCode,
				ShippingStreet:          address.Address,
				ShippingCity:            city,
				ShippingState:           address.StateCode,
				ShippingPostalCode:      address.PostalCode,
				ShippingCountry:         address.CountryCode,
				TotalC:                  order.Total.String(),
				SubTotalC:               order.Subtotal.String(),
				ShippingRateC:           order.ShippingRate.String(),
				TaxAmountC:              order.TaxAmount.String(),
				ShippingCarrierNameC:    order.ShippingCarrierName,
				ShippingCarrierServiceC: order.ShippingServiceType,
				CurrencyC:               order.Currency,
				Pricebook2ID:            salesforceEntities.StandardPriceBook,
				EstimatedDeliveryDateC: func() string {
					if order.ShippingEstimatedDeliveryDate.IsZero() {
						return time.Now().Format("2006-01-02")
					}
					return order.ShippingEstimatedDeliveryDate.Format("2006-01-02")
				}(),
				OrderCreatedAtC: time.Now().Format("2006-01-02"),
			})
			if err != nil {
				s.log.Errorf("Error creating order on salesforce: %v", err)
				return
			}
			if res.Success {
				// update order with salesforce order id
				err = s.repo.Update(bgCtx, map[string]interface{}{
					"salesforce_id": res.ID,
				}, order.ID.String(), order.CustomerID.String())
				if err != nil {
					s.log.Errorf("Error updating order with salesforce order id: %v", err)
					return
				}

				productIDs := func() []string {
					var ids []string
					for _, item := range cartItems.Items {
						ids = append(ids, item.ProductID.String())
					}
					return ids
				}
				// get salesforce products by product ids
				products, err := s.productClient.GetProductsByIDs(bgCtx, productIDs())
				if err != nil {
					s.log.Errorf("Error fetching salesforce products: %v", err)
					return
				}

				sfOrderItems := make([]*salesforceEntities.OrderItem, 0, len(orderItems))

				if len(products) > 0 {
					for _, item := range cartItems.Items {

						description := ""
						if item.Description != nil {
							description = *item.Description
						}

						sfOrderItem := salesforceEntities.OrderItem{
							OrderID:     res.ID,
							Quantity:    item.Quantity,
							UnitPrice:   item.Price.InexactFloat64(),
							Description: item.Name,
							TypeC:       description,
						}

						for _, product := range products {
							if product.ID == item.ProductID && product.SalesforcePricebookEntryId != nil {
								sfOrderItem.PricebookEntryID = *product.SalesforcePricebookEntryId
								break
							}
						}

						sfOrderItems = append(sfOrderItems, &sfOrderItem)
					}

					// add items to order on salesforce
					_, err := s.salesforceClient.AddOrderItems(bgCtx, sfOrderItems)
					if err != nil {
						s.log.Errorf("Error adding items to order on salesforce: %v", err)
						return
					}

					// get order items from salesforce
					items, err := s.salesforceClient.GetOrderItems(bgCtx, res.ID)
					if err != nil {
						s.log.Errorf("Error fetching order items from salesforce: %v", err)
						return
					}

					if len(items.Records) > 0 {
						// map of order item id to salesforce order item id
						ids := make(map[string]string)
						// build the map based on TypeC and Description and ProductID & Product2Id
						// salesforce order items
						for _, item := range items.Records {
							// order items
							for _, orderItem := range orderItems {
								if orderItem.Description != nil &&
									item.TypeC == *orderItem.Description &&
									item.Description == orderItem.Name {
									ids[orderItem.ID.String()] = item.ID
									break
								}
							}
						}

						// update order items with salesforce order item ids
						err = s.repo.AddSalesforceIDPerOrderItem(bgCtx, ids)
						if err != nil {
							s.log.Errorf("Error updating order items with salesforce order item ids: %v", err)
							return
						}
					}
				} else {
					s.log.Errorf("No products found for order items")
				}
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
			ProfileID:    *paymentReq.Customer.AuthorizeNetID,
			PaymentNonce: paymentReq.PaymentNonce,
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

	// update order status on salesforce
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		if customer.SalesforceID != nil && order.SalesforceID != "" {
			err = s.salesforceClient.UpdateOrderStatus(bgCtx, &salesforceEntities.UpdateOrderRequest{
				OrderId:   order.SalesforceID,
				Status:    entities.Cancelled.String(),
				AccountID: *customer.SalesforceID,
			})
			if err != nil {
				s.log.Errorf("Error updating order status on salesforce: %v", err)
				return
			}
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
		err = s.webhookClient.NotifyOrderStatusChange(bgCtx, &webhookEntities.NotifyOrderStatusChangeRequest{
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
		if customer.SalesforceID != nil && order.SalesforceID != "" {
			bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			err = s.salesforceClient.UpdateOrderStatus(bgCtx, &salesforceEntities.UpdateOrderRequest{
				OrderId:   order.SalesforceID,
				Status:    entities.PaymentSuccess.String(),
				AccountID: *customer.SalesforceID,
			})
			if err != nil {
				s.log.Errorf("Error updating order status on salesforce: %v", err)
				return
			}
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
		err = s.webhookClient.NotifyOrderStatusChange(bgCtx, &webhookEntities.NotifyOrderStatusChangeRequest{
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
		if customer.SalesforceID != nil && order.SalesforceID != "" {
			bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			err = s.salesforceClient.UpdateOrderStatus(bgCtx, &salesforceEntities.UpdateOrderRequest{
				OrderId:   order.SalesforceID,
				Status:    entities.PaymentFailed.String(),
				AccountID: *customer.SalesforceID,
			})
			if err != nil {
				s.log.Errorf("Error updating order status on salesforce: %v", err)
				return
			}
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

			err = s.webhookClient.NotifyOrderStatusChange(bgCtx, &webhookEntities.NotifyOrderStatusChangeRequest{
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

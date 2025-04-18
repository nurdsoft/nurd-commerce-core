package service

import (
	"context"
	"time"

	"github.com/brianvoe/gofakeit"
	userBookClient "github.com/nurdsoft/nurd-commerce-core/internal/userbook/client"
	"github.com/shopspring/decimal"

	"testing"

	sfEntities "github.com/nurdsoft/nurd-commerce-core/internal/vendors/salesforce/entities"

	salesforce "github.com/nurdsoft/nurd-commerce-core/internal/vendors/salesforce/client"

	stripeClient "github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/client"
	stripeEntities "github.com/nurdsoft/nurd-commerce-core/internal/vendors/stripe/entities"
	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	cartClient "github.com/nurdsoft/nurd-commerce-core/internal/cart/cartclient"
	cartEntities "github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/orders/repository"
	userEntities "github.com/nurdsoft/nurd-commerce-core/internal/user/entities"
	userClient "github.com/nurdsoft/nurd-commerce-core/internal/user/userclient"
	sesClient "github.com/nurdsoft/nurd-commerce-core/internal/vendors/aws/ses/client"
	cdfliteClient "github.com/nurdsoft/nurd-commerce-core/internal/vendors/cdflite/client"
	cdfliteEntities "github.com/nurdsoft/nurd-commerce-core/internal/vendors/cdflite/entities"
	wishlistClient "github.com/nurdsoft/nurd-commerce-core/internal/wishlist/wishlistclient"
	"github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_CreateOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSfId := "sf_user_id"

	setup := func() (
		*service, context.Context,
		*repository.MockRepository,
		*userClient.MockClient,
		*cartClient.MockClient,
		*stripeClient.MockClient,
		*salesforce.MockClient,
	) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockUserClient := userClient.NewMockClient(ctrl)
		mockCartClient := cartClient.NewMockClient(ctrl)
		mockStripeClient := stripeClient.NewMockClient(ctrl)
		mockUserBookClient := userBookClient.NewMockClient(ctrl)
		mockSfClient := salesforce.NewMockClient(ctrl)
		userUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo:             mockRepo,
			log:              zap.NewExample().Sugar(),
			userClient:       mockUserClient,
			cartClient:       mockCartClient,
			stripeClient:     mockStripeClient,
			userBookClient:   mockUserBookClient,
			salesforceClient: mockSfClient,
		}
		return svc, ctx, mockRepo, mockUserClient, mockCartClient, mockStripeClient, mockSfClient
	}

	t.Run("CreateOrder", func(t *testing.T) {
		svc, ctx, mockRepo, mockUserClient, mockCartClient, mockStripeClient, mockSfClient := setup()
		req := &entities.CreateOrderRequest{
			Body: &entities.CreateOrderRequestBody{
				AddressID: uuid.New(),
			},
		}

		mockUserClient.EXPECT().GetAddressByUUID(ctx, &userEntities.GetAddressRequest{
			AddressUUID: req.Body.AddressID,
		}).Return(
			&userEntities.GetAddressResponse{
				Address: userEntities.UserAddress{
					City:    "City",
					State:   "State",
					Country: "US",
					ZipCode: "12345",
				},
			}, nil,
		)

		mockCartClient.EXPECT().GetCart(ctx).Return(&cartEntities.Cart{}, nil)
		mockCartClient.EXPECT().GetCartItems(ctx).Return(&cartEntities.GetCartItemsResponse{}, nil)
		mockCartClient.EXPECT().GetShippingRateByUUID(ctx, gomock.Any()).Return(&cartEntities.CartShippingRate{}, nil)
		mockUserClient.EXPECT().GetUser(ctx).Return(&userEntities.User{
			SFUserID: &mockSfId,
		}, nil)
		mockStripeClient.EXPECT().CreatePaymentIntent(ctx, gomock.Any()).Return(&stripeEntities.CreatePaymentIntentResponse{}, nil)
		mockRepo.EXPECT().CreateOrder(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockSfClient.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).Return(&sfEntities.CreateSFOrderResponse{}, nil).AnyTimes()
		mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockCartClient.EXPECT().GetSalesforceProudctsByBookIDs(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockSfClient.EXPECT().AddOrderItems(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockSfClient.EXPECT().GetOrderItems(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockRepo.EXPECT().AddSFIDPerOrderItem(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockRepo.EXPECT().OrderReferenceExists(gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()

		_, err := svc.CreateOrder(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("CreateOrder with an existing order reference", func(t *testing.T) {
		svc, ctx, mockRepo, mockUserClient, mockCartClient, mockStripeClient, mockSfClient := setup()
		req := &entities.CreateOrderRequest{
			Body: &entities.CreateOrderRequestBody{
				AddressID: uuid.New(),
			},
		}

		mockUserClient.EXPECT().GetAddressByUUID(ctx, &userEntities.GetAddressRequest{
			AddressUUID: req.Body.AddressID,
		}).Return(
			&userEntities.GetAddressResponse{
				Address: userEntities.UserAddress{
					City:    "City",
					State:   "State",
					Country: "US",
					ZipCode: "12345",
				},
			}, nil,
		)

		mockCartClient.EXPECT().GetCart(ctx).Return(&cartEntities.Cart{}, nil)
		mockCartClient.EXPECT().GetCartItems(ctx).Return(&cartEntities.GetCartItemsResponse{}, nil)
		mockCartClient.EXPECT().GetShippingRateByUUID(ctx, gomock.Any()).Return(&cartEntities.CartShippingRate{}, nil)
		mockUserClient.EXPECT().GetUser(ctx).Return(&userEntities.User{
			SFUserID: &mockSfId,
		}, nil)
		mockStripeClient.EXPECT().CreatePaymentIntent(ctx, gomock.Any()).Return(&stripeEntities.CreatePaymentIntentResponse{}, nil)
		mockRepo.EXPECT().CreateOrder(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockSfClient.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).Return(&sfEntities.CreateSFOrderResponse{}, nil).AnyTimes()
		mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		mockCartClient.EXPECT().GetSalesforceProudctsByBookIDs(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockSfClient.EXPECT().AddOrderItems(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockSfClient.EXPECT().GetOrderItems(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockRepo.EXPECT().AddSFIDPerOrderItem(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		gomock.InOrder(
			mockRepo.EXPECT().OrderReferenceExists(ctx, gomock.Any()).Return(true, nil),
			mockRepo.EXPECT().OrderReferenceExists(ctx, gomock.Any()).Return(false, nil),
		)
		_, err := svc.CreateOrder(ctx, req)
		assert.NoError(t, err)
	})
}

func Test_ListOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository, *userClient.MockClient, *cartClient.MockClient) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockUserClient := userClient.NewMockClient(ctrl)
		mockCartClient := cartClient.NewMockClient(ctrl)
		userUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo:       mockRepo,
			log:        zap.NewExample().Sugar(),
			userClient: mockUserClient,
			cartClient: mockCartClient,
		}
		return svc, ctx, mockRepo, mockUserClient, mockCartClient
	}

	t.Run("ListOrders", func(t *testing.T) {
		svc, ctx, mockRepo, _, _ := setup()
		req := &entities.ListOrdersRequest{
			Limit: 10,
		}

		mockRepo.EXPECT().ListOrders(ctx, gomock.Any(), req.Limit, "").Return([]*entities.Order{}, "", nil)
		_, err := svc.ListOrders(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("ListOrdersWithCursor", func(t *testing.T) {
		svc, ctx, mockRepo, _, _ := setup()
		req := &entities.ListOrdersRequest{
			Limit:  10,
			Cursor: "cursor",
		}

		mockRepo.EXPECT().ListOrders(ctx, gomock.Any(), req.Limit, req.Cursor).Return([]*entities.Order{}, "", nil)
		_, err := svc.ListOrders(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("ListOrdersWithInvalidCursor", func(t *testing.T) {
		svc, ctx, mockRepo, _, _ := setup()
		req := &entities.ListOrdersRequest{
			Limit:  10,
			Cursor: "invalid",
		}

		mockRepo.EXPECT().ListOrders(ctx, gomock.Any(), req.Limit, req.Cursor).Return(nil, "", assert.AnError)
		_, err := svc.ListOrders(ctx, req)
		assert.Error(t, err)
	})

	t.Run("ListOrdersWithInvalidLimit", func(t *testing.T) {
		svc, ctx, mockRepo, _, _ := setup()
		req := &entities.ListOrdersRequest{
			Limit:  0,
			Cursor: "",
		}

		mockRepo.EXPECT().ListOrders(ctx, gomock.Any(), req.Limit, req.Cursor).Return(nil, "", assert.AnError)
		_, err := svc.ListOrders(ctx, req)
		assert.Error(t, err)
	})
}

func Test_service_ProcessPaymentSucceeded(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository, *userClient.MockClient,
		*sesClient.MockClient, *userBookClient.MockClient, *wishlistClient.MockClient, *cdfliteClient.MockClient) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockUserClient := userClient.NewMockClient(ctrl)
		mockSesClient := sesClient.NewMockClient(ctrl)
		mockUserBookClient := userBookClient.NewMockClient(ctrl)
		mockWishlistClient := wishlistClient.NewMockClient(ctrl)
		mockCdfliteClient := cdfliteClient.NewMockClient(ctrl)
		ctx := context.Background()
		svc := &service{
			repo:           mockRepo,
			log:            zap.NewExample().Sugar(),
			userClient:     mockUserClient,
			sesClient:      mockSesClient,
			userBookClient: mockUserBookClient,
			wishlistClient: mockWishlistClient,
			cdfliteClient:  mockCdfliteClient,
		}
		return svc, ctx, mockRepo, mockUserClient, mockSesClient, mockUserBookClient, mockWishlistClient, mockCdfliteClient
	}

	t.Run("order not found", func(t *testing.T) {
		svc, ctx, mockRepo, _, _, _, _, _ := setup()
		paymentIntentId := "pi_123"

		mockRepo.EXPECT().GetOrderByPaymentIntentId(ctx, paymentIntentId).Return(nil, &appErrors.APIError{Message: "order not found"})

		err := svc.ProcessPaymentSucceeded(ctx, paymentIntentId)
		assert.IsType(t, &appErrors.APIError{}, err)
	})

	t.Run("order is not pending", func(t *testing.T) {
		svc, ctx, mockRepo, _, _, _, _, _ := setup()
		paymentIntentId := "pi_123"
		order := &entities.Order{
			ID:     uuid.New(),
			Status: entities.PaymentFailed,
		}

		mockRepo.EXPECT().GetOrderByPaymentIntentId(ctx, paymentIntentId).Return(order, nil)

		err := svc.ProcessPaymentSucceeded(ctx, paymentIntentId)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Contains(t, err.Error(), "Order is not pending.")
	})

	t.Run("repository update fails", func(t *testing.T) {
		svc, ctx, mockRepo, _, _, _, _, _ := setup()
		paymentIntentId := "pi_123"
		order := &entities.Order{
			ID:     uuid.New(),
			Status: entities.Pending,
		}

		mockRepo.EXPECT().GetOrderByPaymentIntentId(ctx, paymentIntentId).Return(order, nil)
		mockRepo.EXPECT().Update(ctx, map[string]interface{}{
			"status": entities.PaymentSuccess,
		}, order.ID.String(), order.UserUUID.String()).Return(&appErrors.APIError{Message: "update failed"})

		err := svc.ProcessPaymentSucceeded(ctx, paymentIntentId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update failed")
	})

	t.Run("successfully processes payment succeeded", func(t *testing.T) {
		svc, ctx, mockRepo, mockUserClient, mockSesClient, mockUserBookClient, mockWishlistClient, mockCdfliteClient := setup()
		paymentIntentId := "pi_123"
		userUUID := uuid.New()
		order := &entities.Order{
			ID:       uuid.New(),
			Status:   entities.Pending,
			UserUUID: userUUID,
		}

		mockUserClient.EXPECT().GetUserByUUID(ctx, userUUID).Return(
			&userEntities.User{
				FirstName: "John",
				Email:     "john@doe.com",
			}, nil,
		)

		mockRepo.EXPECT().GetOrderByPaymentIntentId(ctx, paymentIntentId).Return(order, nil)
		mockRepo.EXPECT().GetOrderItemsByOrderID(ctx, order.ID).Return([]*entities.OrderItem{}, nil)
		mockWishlistClient.EXPECT().RemoveBooksFromWishlist(ctx, gomock.Any()).Return(nil)
		mockSesClient.EXPECT().SendEmail(ctx, gomock.Any()).Return(nil)
		mockUserBookClient.EXPECT().AddUserBooks(ctx, gomock.Any(), gomock.Any()).Return(nil)
		mockRepo.EXPECT().Update(ctx, map[string]interface{}{
			"status": entities.PaymentSuccess,
		}, order.ID.String(), order.UserUUID.String()).Return(nil)
		mockCdfliteClient.EXPECT().SubmitPurchaseOrder(ctx, gomock.Any()).Return(&cdfliteEntities.OrderResponse{Message: "CDFLite successful message"}, nil)
		mockRepo.EXPECT().Update(ctx, map[string]interface{}{
			"fulfillment_vendor_message": "CDFLite successful message",
		}, order.ID.String(), order.UserUUID.String()).Return(nil)
		err := svc.ProcessPaymentSucceeded(ctx, paymentIntentId)
		assert.NoError(t, err)
	})
}

func Test_service_ProcessPaymentFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository, *userClient.MockClient, *sesClient.MockClient) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockUserClient := userClient.NewMockClient(ctrl)
		mockSesClient := sesClient.NewMockClient(ctrl)
		ctx := context.Background()
		svc := &service{
			repo:       mockRepo,
			log:        zap.NewExample().Sugar(),
			userClient: mockUserClient,
			sesClient:  mockSesClient,
		}
		return svc, ctx, mockRepo, mockUserClient, mockSesClient
	}

	t.Run("order not found", func(t *testing.T) {
		svc, ctx, mockRepo, _, _ := setup()
		paymentIntentId := "pi_123"

		mockRepo.EXPECT().GetOrderByPaymentIntentId(ctx, paymentIntentId).Return(nil, &appErrors.APIError{Message: "order not found"})

		err := svc.ProcessPaymentFailed(ctx, paymentIntentId)
		assert.IsType(t, &appErrors.APIError{}, err)
	})

	t.Run("order is not pending", func(t *testing.T) {
		svc, ctx, mockRepo, _, _ := setup()
		paymentIntentId := "pi_123"
		order := &entities.Order{
			ID:     uuid.New(),
			Status: entities.PaymentSuccess,
		}

		mockRepo.EXPECT().GetOrderByPaymentIntentId(ctx, paymentIntentId).Return(order, nil)

		err := svc.ProcessPaymentFailed(ctx, paymentIntentId)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Contains(t, err.Error(), "Order is not pending.")
	})

	t.Run("repository update fails", func(t *testing.T) {
		svc, ctx, mockRepo, _, _ := setup()
		paymentIntentId := "pi_123"
		order := &entities.Order{
			ID:     uuid.New(),
			Status: entities.Pending,
		}

		mockRepo.EXPECT().GetOrderByPaymentIntentId(ctx, paymentIntentId).Return(order, nil)
		mockRepo.EXPECT().Update(ctx, map[string]interface{}{
			"status": entities.PaymentFailed,
		}, order.ID.String(), order.UserUUID.String()).Return(&appErrors.APIError{Message: "update failed"})

		err := svc.ProcessPaymentFailed(ctx, paymentIntentId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update failed")
	})

	t.Run("successfully processes payment failed", func(t *testing.T) {
		svc, ctx, mockRepo, mockUserClient, mockSesClient := setup()
		paymentIntentId := "pi_123"
		userUUID := uuid.New()
		order := &entities.Order{
			ID:       uuid.New(),
			Status:   entities.Pending,
			UserUUID: userUUID,
		}

		mockUserClient.EXPECT().GetUserByUUID(ctx, userUUID).Return(
			&userEntities.User{
				FirstName: "John",
				Email:     "john@doe.com",
			}, nil,
		)
		mockRepo.EXPECT().GetOrderByPaymentIntentId(ctx, paymentIntentId).Return(order, nil)
		mockSesClient.EXPECT().SendEmail(ctx, gomock.Any()).Return(nil)

		mockRepo.EXPECT().Update(ctx, map[string]interface{}{
			"status": entities.PaymentFailed,
		}, order.ID.String(), order.UserUUID.String()).Return(nil)

		err := svc.ProcessPaymentFailed(ctx, paymentIntentId)
		assert.NoError(t, err)
	})
}

func Test_service_GetOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		userUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, ctx, mockRepo
	}

	t.Run("unable to retrieve order", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		req := &entities.GetOrderRequest{
			OrderID: uuid.New(),
		}

		mockRepo.EXPECT().GetOrderByUUID(ctx, req.OrderID).Return(nil, assert.AnError)

		_, err := svc.GetOrder(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})

	t.Run("unable to retrieve order items", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		orderID := uuid.New()
		userId, _ := uuid.Parse(meta.XCustomerID(ctx))
		req := &entities.GetOrderRequest{
			OrderID: orderID,
		}
		order := &entities.Order{
			ID:       orderID,
			UserUUID: userId,
		}

		mockRepo.EXPECT().GetOrderByUUID(ctx, req.OrderID).Return(order, nil)
		mockRepo.EXPECT().GetOrderItemsByOrderID(ctx, order.ID).Return(nil, assert.AnError)

		_, err := svc.GetOrder(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})

	t.Run("successfully retrieves order", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		orderID := uuid.New()
		userId, _ := uuid.Parse(meta.XCustomerID(ctx))

		req := &entities.GetOrderRequest{
			OrderID: orderID,
		}
		order := &entities.Order{
			ID:       orderID,
			UserUUID: userId,
		}
		orderItems := []*entities.OrderItem{
			{
				ID: uuid.New(),
			},
		}

		mockRepo.EXPECT().GetOrderByUUID(ctx, req.OrderID).Return(order, nil)
		mockRepo.EXPECT().GetOrderItemsByOrderID(ctx, order.ID).Return(orderItems, nil)

		resp, err := svc.GetOrder(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, order, resp.Order)
		assert.Equal(t, orderItems, resp.OrderItems)
	})

	t.Run("order does not belong to user", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		orderID := uuid.New()
		req := &entities.GetOrderRequest{
			OrderID: orderID,
		}
		order := &entities.Order{
			ID:       orderID,
			UserUUID: uuid.New(),
		}

		mockRepo.EXPECT().GetOrderByUUID(ctx, req.OrderID).Return(order, nil)

		_, err := svc.GetOrder(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})
}

func Test_service_CancelOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository, *userClient.MockClient,
		*sesClient.MockClient, *userBookClient.MockClient) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockUserClient := userClient.NewMockClient(ctrl)
		mockSesClient := sesClient.NewMockClient(ctrl)
		mockUserBookClient := userBookClient.NewMockClient(ctrl)
		userUUID := uuid.New()
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo:           mockRepo,
			log:            zap.NewExample().Sugar(),
			userClient:     mockUserClient,
			sesClient:      mockSesClient,
			userBookClient: mockUserBookClient,
		}
		return svc, ctx, mockRepo, mockUserClient, mockSesClient, mockUserBookClient
	}

	t.Run("unable to retrieve order", func(t *testing.T) {
		svc, ctx, mockRepo, _, _, _ := setup()
		req := &entities.CancelOrderRequest{
			OrderID: uuid.New(),
		}

		mockRepo.EXPECT().GetOrderByUUID(ctx, req.OrderID).Return(nil, assert.AnError)

		err := svc.CancelOrder(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})

	t.Run("order does not belong to user", func(t *testing.T) {
		svc, ctx, mockRepo, _, _, _ := setup()
		orderID := uuid.New()
		req := &entities.CancelOrderRequest{
			OrderID: orderID,
		}
		order := &entities.Order{
			ID:       orderID,
			UserUUID: uuid.New(),
		}

		mockRepo.EXPECT().GetOrderByUUID(ctx, req.OrderID).Return(order, nil)

		err := svc.CancelOrder(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})

	t.Run("order is not cancellable", func(t *testing.T) {
		svc, ctx, mockRepo, _, _, _ := setup()
		orderID := uuid.New()
		req := &entities.CancelOrderRequest{
			OrderID: orderID,
		}
		order := &entities.Order{
			ID:       orderID,
			UserUUID: uuid.New(),
			Status:   entities.PaymentSuccess,
		}

		mockRepo.EXPECT().GetOrderByUUID(ctx, req.OrderID).Return(order, nil)

		err := svc.CancelOrder(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})

	t.Run("successfully cancels order", func(t *testing.T) {
		svc, ctx, mockRepo, mockUserClient, mockSesClient, mockUserBookClient := setup()
		orderID := uuid.New()
		userUUID, _ := uuid.Parse(meta.XCustomerID(ctx))
		req := &entities.CancelOrderRequest{
			OrderID: orderID,
		}
		order := &entities.Order{
			ID:       orderID,
			UserUUID: userUUID,
			Status:   entities.Pending,
		}

		mockRepo.EXPECT().GetOrderByUUID(ctx, req.OrderID).Return(order, nil)
		mockUserClient.EXPECT().GetUserByUUID(ctx, userUUID).Return(
			&userEntities.User{
				FirstName: "John",
				Email:     "joh@doe.com",
			}, nil,
		)
		mockRepo.EXPECT().GetOrderItemsByOrderID(ctx, order.ID).Return([]*entities.OrderItem{}, nil)
		mockUserBookClient.EXPECT().DeletePurchasedUserBook(ctx, gomock.Any(), gomock.Any()).Return(nil)
		mockSesClient.EXPECT().SendEmail(ctx, gomock.Any()).Return(nil)
		mockRepo.EXPECT().Update(ctx, map[string]interface{}{
			"status": entities.Cancelled,
		}, order.ID.String(), order.UserUUID.String()).Return(nil)

		err := svc.CancelOrder(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("order is already cancelled", func(t *testing.T) {
		svc, ctx, mockRepo, _, _, _ := setup()
		orderID := uuid.New()
		userUUID, _ := uuid.Parse(meta.XCustomerID(ctx))
		req := &entities.CancelOrderRequest{
			OrderID: orderID,
		}
		order := &entities.Order{
			ID:       orderID,
			UserUUID: userUUID,
			Status:   entities.Cancelled,
		}

		mockRepo.EXPECT().GetOrderByUUID(ctx, req.OrderID).Return(order, nil)

		err := svc.CancelOrder(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})
}

func Test_service_ProcessOrderShipped(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository, *userClient.MockClient,
		*sesClient.MockClient, *userBookClient.MockClient, *salesforce.MockClient) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockUserClient := userClient.NewMockClient(ctrl)
		mockSesClient := sesClient.NewMockClient(ctrl)
		mockUserBookClient := userBookClient.NewMockClient(ctrl)
		mockSfClient := salesforce.NewMockClient(ctrl)
		ctx := context.Background()
		svc := &service{
			repo:             mockRepo,
			log:              zap.NewExample().Sugar(),
			userClient:       mockUserClient,
			sesClient:        mockSesClient,
			userBookClient:   mockUserBookClient,
			salesforceClient: mockSfClient,
		}
		return svc, ctx, mockRepo, mockUserClient, mockSesClient, mockUserBookClient, mockSfClient
	}
	t.Run("order not found", func(t *testing.T) {
		svc, ctx, mockRepo, _, _, _, _ := setup()
		orderID := gofakeit.Word()
		req := &entities.IngramOrderFullfillment{
			ShipmentDate:  time.Time{},
			FreightCharge: decimal.Decimal{},
			OrderTotal:    decimal.Decimal{},
		}

		mockRepo.EXPECT().GetOrderByReference(ctx, orderID).Return(nil, &appErrors.APIError{Message: "order not found"})

		err := svc.ProcessOrderShipped(ctx, orderID, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})

	t.Run("repository update fails", func(t *testing.T) {
		svc, ctx, mockRepo, _, _, _, _ := setup()
		orderRef := gofakeit.Word()
		orderID := uuid.New()
		userId := uuid.New()
		req := &entities.IngramOrderFullfillment{
			ShipmentDate:  time.Time{},
			FreightCharge: decimal.Decimal{},
			OrderTotal:    decimal.Decimal{},
		}
		order := &entities.Order{
			ID:       orderID,
			Status:   entities.Pending,
			UserUUID: userId,
		}

		mockRepo.EXPECT().GetOrderByReference(ctx, orderRef).Return(order, nil)
		mockRepo.EXPECT().Update(ctx, map[string]interface{}{
			"status":                            entities.Shipped,
			"fulfillment_vendor_freight_charge": req.FreightCharge,
			"fulfillment_vendor_order_total":    req.OrderTotal,
			"fulfillment_vendor_shipment_date":  req.ShipmentDate,
		}, order.ID.String(), order.UserUUID.String()).Return(&appErrors.APIError{Message: "update failed"})

		err := svc.ProcessOrderShipped(ctx, orderRef, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update failed")
	})

	t.Run("successfully processes order shipped", func(t *testing.T) {
		svc, ctx, mockRepo, _, _, _, _ := setup()
		orderRef := gofakeit.Word()
		orderID := uuid.New()
		userId := uuid.New()
		req := &entities.IngramOrderFullfillment{
			ShipmentDate:  time.Now(),
			FreightCharge: decimal.NewFromFloat(10.0),
			OrderTotal:    decimal.NewFromFloat(100.0),
		}
		order := &entities.Order{
			ID:       orderID,
			UserUUID: userId,
			Status:   entities.Pending,
		}

		mockRepo.EXPECT().GetOrderByReference(ctx, orderRef).Return(order, nil)
		mockRepo.EXPECT().Update(ctx, map[string]interface{}{
			"status":                            entities.Shipped,
			"fulfillment_vendor_freight_charge": req.FreightCharge,
			"fulfillment_vendor_order_total":    req.OrderTotal,
			"fulfillment_vendor_shipment_date":  req.ShipmentDate,
		}, order.ID.String(), order.UserUUID.String()).Return(nil)

		err := svc.ProcessOrderShipped(ctx, orderRef, req)
		assert.NoError(t, err)

	})
}

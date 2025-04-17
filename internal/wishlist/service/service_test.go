package service

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	cartclient "github.com/nurdsoft/nurd-commerce-core/internal/cart/cartclient"
	cartEntities "github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/wishlist/repository"
	appErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_service_AddToWishlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		userUUID := uuid.New()
		ctx := meta.XCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, ctx, mockRepo
	}

	t.Run("Valid request with a book id", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		req := &entities.AddToWishlistRequest{
			Body: &entities.AddToWishlistRequestBody{
				BookUUIDs: []uuid.UUID{uuid.New()},
			},
		}

		mockRepo.EXPECT().UpdateWishlist(ctx, meta.XCustomerID(ctx), req.Body.BookUUIDs).Return(nil).Times(1)
		err := svc.AddToWishlist(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("Add the same book to wishlist twice", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		bookUUID := uuid.New()
		req := &entities.AddToWishlistRequest{
			Body: &entities.AddToWishlistRequestBody{
				BookUUIDs: []uuid.UUID{bookUUID},
			},
		}

		// First addition should succeed
		mockRepo.EXPECT().UpdateWishlist(ctx, meta.XCustomerID(ctx), req.Body.BookUUIDs).Return(nil).Times(1)
		err := svc.AddToWishlist(ctx, req)
		assert.NoError(t, err)

		// Second addition should not fail with an error
		mockRepo.EXPECT().UpdateWishlist(ctx, meta.XCustomerID(ctx), req.Body.BookUUIDs).Return(nil).Times(1)
		err = svc.AddToWishlist(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("no user ID", func(t *testing.T) {
		svc, _, _ := setup()
		req := &entities.AddToWishlistRequest{
			Body: &entities.AddToWishlistRequestBody{
				BookUUIDs: []uuid.UUID{uuid.New()},
			},
		}
		ctx := meta.XCustomerID(context.Background(), "")
		err := svc.AddToWishlist(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})
}

func Test_service_RemoveFromWishlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		userUUID := uuid.New()
		ctx := meta.XCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, ctx, mockRepo
	}

	t.Run("Valid request with a book id", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		req := &entities.RemoveFromWishlistRequest{
			BookUUID: uuid.New(),
		}

		mockRepo.EXPECT().DeleteBookFromWishlist(ctx, meta.XCustomerID(ctx), req.BookUUID).Return(nil).Times(1)
		err := svc.RemoveFromWishlist(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("Remove the same book from wishlist twice", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		bookUUID := uuid.New()
		req := &entities.RemoveFromWishlistRequest{
			BookUUID: bookUUID,
		}

		// First removal should succeed
		mockRepo.EXPECT().DeleteBookFromWishlist(ctx, meta.XCustomerID(ctx), req.BookUUID).Return(nil).Times(1)
		err := svc.RemoveFromWishlist(ctx, req)
		assert.NoError(t, err)

		// Second removal should return not found error
		mockRepo.EXPECT().DeleteBookFromWishlist(ctx, meta.XCustomerID(ctx), req.BookUUID).Return(&appErrors.APIError{}).Times(1)
		err = svc.RemoveFromWishlist(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})

	t.Run("no user ID", func(t *testing.T) {
		svc, _, _ := setup()
		req := &entities.RemoveFromWishlistRequest{
			BookUUID: uuid.New(),
		}
		ctx := meta.XCustomerID(context.Background(), "")
		err := svc.RemoveFromWishlist(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
	})
}

func Test_service_GetWishlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		userUUID := uuid.New()
		ctx := meta.XCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, ctx, mockRepo
	}

	t.Run("GetWishlistWithNoUserID", func(t *testing.T) {
		svc, _, _ := setup()
		req := &entities.GetWishlistRequest{}
		ctx := meta.XCustomerID(context.Background(), "")
		wishlist, err := svc.GetWishlist(ctx, req)
		assert.IsType(t, &appErrors.APIError{}, err)
		assert.Nil(t, wishlist)
	})

	t.Run("GetWishlist", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		req := &entities.GetWishlistRequest{
			Limit: 10,
		}
		wishlist := []*entities.Wishlist{
			{
				Id:       uuid.New(),
				UserUUID: uuid.New(),
				BookID:   uuid.New(),
			},
		}
		mockRepo.EXPECT().GetWishlist(ctx, meta.XCustomerID(ctx), req.Limit, "").Return(wishlist, "", nil).Times(1)
		_, err := svc.GetWishlist(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("GetWishlistWithCursor", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		req := &entities.GetWishlistRequest{
			Limit:  10,
			Cursor: "cursor",
		}
		wishlist := []*entities.Wishlist{
			{
				Id:       uuid.New(),
				UserUUID: uuid.New(),
				BookID:   uuid.New(),
			},
		}
		mockRepo.EXPECT().GetWishlist(ctx, meta.XCustomerID(ctx), req.Limit, req.Cursor).Return(wishlist, "", nil).Times(1)
		_, err := svc.GetWishlist(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("GetWishlistWithInvalidCursor", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		req := &entities.GetWishlistRequest{
			Limit:  10,
			Cursor: "invalid_cursor",
		}
		mockRepo.EXPECT().GetWishlist(ctx, meta.XCustomerID(ctx), req.Limit, req.Cursor).Return(nil, "", assert.AnError).Times(1)
		_, err := svc.GetWishlist(ctx, req)
		assert.Error(t, err)
	})

	t.Run("GetWishlistWithInvalidLimit", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		req := &entities.GetWishlistRequest{
			Limit:  0,
			Cursor: "cursor",
		}
		mockRepo.EXPECT().GetWishlist(ctx, meta.XCustomerID(ctx), req.Limit, req.Cursor).Return(nil, "", assert.AnError).Times(1)
		_, err := svc.GetWishlist(ctx, req)
		assert.Error(t, err)
	})
}

func Test_service_RemoveBooksFromWishlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		userUUID := uuid.New()
		ctx := meta.XCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo: mockRepo,
			log:  zap.NewExample().Sugar(),
		}
		return svc, ctx, mockRepo
	}

	t.Run("successfully removes books from wishlist", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		userUUID := uuid.New()
		req := &entities.BulkRemoveFromWishlistRequest{
			UserUUID:  userUUID,
			BookUUIDs: []uuid.UUID{uuid.New(), uuid.New()},
		}

		mockRepo.EXPECT().BulkRemoveFromWishlist(ctx, req.UserUUID, req.BookUUIDs).Return(nil).Times(1)

		err := svc.RemoveBooksFromWishlist(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("fails to remove books from wishlist", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		userUUID := uuid.New()
		req := &entities.BulkRemoveFromWishlistRequest{
			UserUUID:  userUUID,
			BookUUIDs: []uuid.UUID{uuid.New(), uuid.New()},
		}

		mockRepo.EXPECT().BulkRemoveFromWishlist(ctx, req.UserUUID, req.BookUUIDs).Return(&appErrors.APIError{Message: "database error"}).Times(1)

		err := svc.RemoveBooksFromWishlist(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
	})
}

func Test_service_GetMoreFromWishlist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userUUID := uuid.New()
	cartBookId := uuid.New()
	wishListBookId1 := cartBookId
	wishListBookId2 := uuid.New()

	setup := func() (*service, context.Context, *repository.MockRepository, *cartclient.MockClient) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockCartClient := cartclient.NewMockClient(ctrl)
		ctx := meta.XCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo:       mockRepo,
			log:        zap.NewExample().Sugar(),
			cartClient: mockCartClient,
		}
		return svc, ctx, mockRepo, mockCartClient
	}

	t.Run("successfully gets more books from wishlist", func(t *testing.T) {
		svc, ctx, mockRepo, mockCartClient := setup()
		req := &entities.GetMoreFromWishlistRequest{}

		mockCartClient.EXPECT().GetCartItems(ctx).Return(&cartEntities.GetCartItemsResponse{
			Items: []cartEntities.CartItem{
				{
					BookId:   cartBookId,
					Quantity: 1,
				},
			},
		}, nil).Times(1)
		mockRepo.EXPECT().GetWishlist(ctx, userUUID.String(), 0, "").Return([]*entities.Wishlist{
			{
				Id:       uuid.New(),
				UserUUID: userUUID,
				BookID:   wishListBookId1,
			},
			{
				Id:       uuid.New(),
				UserUUID: userUUID,
				BookID:   wishListBookId2,
			},
		}, "", nil).Times(1)

		resp, err := svc.GetMoreFromWishlist(ctx, req)
		assert.NoError(t, err)
		assert.Len(t, resp.Items, 1)
		assert.Equal(t, wishListBookId2, resp.Items[0].BookUUID)
	})

	t.Run("fails to get more books from wishlist", func(t *testing.T) {
		svc, ctx, mockRepo, mockCartClient := setup()
		req := &entities.GetMoreFromWishlistRequest{}

		mockCartClient.EXPECT().GetCartItems(ctx).Return(nil, nil).Times(1)
		mockRepo.EXPECT().GetWishlist(ctx, userUUID.String(), 0, "").Return(nil, "", &appErrors.APIError{Message: "database error"}).Times(1)

		_, err := svc.GetMoreFromWishlist(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &appErrors.APIError{}, err)
	})
}

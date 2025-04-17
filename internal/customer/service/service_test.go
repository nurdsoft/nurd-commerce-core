package service

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/internal/user/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/user/repository"
	salesforce "github.com/nurdsoft/nurd-commerce-core/internal/vendors/salesforce/client"
	sfEntities "github.com/nurdsoft/nurd-commerce-core/internal/vendors/salesforce/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/cfg"
	"github.com/nurdsoft/nurd-commerce-core/shared/meta"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_service_GetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUUID := uuid.New()

	setup := func() (*service, context.Context, *repository.MockRepository) {
		mockRepo := repository.NewMockRepository(ctrl)
		ctx := meta.WithXCustomerID(context.Background(), userUUID.String())
		svc := &service{
			repo:   mockRepo,
			log:    zap.NewExample().Sugar(),
			config: cfg.Config{CDNDomain: "cdn.bookstore.ai"},
		}
		return svc, ctx, mockRepo
	}

	t.Run("user not found in request", func(t *testing.T) {
		svc, _, _ := setup()
		_, err := svc.GetUser(context.Background())
		assert.Error(t, err)
	})

	t.Run("user found with empty profile pic", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		mockRepo.EXPECT().FindByUUID(ctx, userUUID.String()).Return(&entities.User{UserUUID: userUUID}, nil)

		user, err := svc.GetUser(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userUUID, user.UserUUID)
		assert.Equal(t, "", user.ProfilePic)
	})

	t.Run("user found with non-empty profile pic", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		mockRepo.EXPECT().FindByUUID(ctx, userUUID.String()).Return(&entities.User{UserUUID: userUUID, ProfilePic: "profile-pic"}, nil)

		user, err := svc.GetUser(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "https://cdn.bookstore.ai/profile-pic", user.ProfilePic)
	})

	t.Run("repository error", func(t *testing.T) {
		svc, ctx, mockRepo := setup()
		mockRepo.EXPECT().FindByUUID(ctx, userUUID.String()).Return(nil, errors.New("repository error"))

		_, err := svc.GetUser(ctx)
		assert.Error(t, err)
	})
}

func Test_service_CreateUser(t *testing.T) {
	type fields struct {
		repo             repository.Repository
		log              *zap.SugaredLogger
		salesforceClient salesforce.Client
	}
	type args struct {
		ctx context.Context
		req *entities.CreateUserRequest
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockRepository(ctrl)
	mockSFClient := salesforce.NewMockClient(ctrl)

	userId := uuid.New()
	email := gofakeit.Email()
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()
	iamUID := uuid.New()
	iamProvider := gofakeit.Word()
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "CreateUser",
			fields: fields{
				repo:             mockRepo,
				log:              zap.NewExample().Sugar(),
				salesforceClient: mockSFClient,
			},
			args: args{
				ctx: context.Background(),
				req: &entities.CreateUserRequest{
					FirstName: firstName,
					LastName:  &lastName,
					Email:     email,
					IAMId:     iamUID.String(),
					IAM:       entities.IAMProvider(iamProvider),
				},
			},
			want: &entities.User{
				UserUUID:                 userId,
				FirstName:                firstName,
				LastName:                 &lastName,
				Email:                    email,
				IAMUID:                   iamUID.String(),
				IAMProvider:              entities.IAMProvider(iamProvider),
				Status:                   "new",
				HasSwipedRecommendations: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				repo:             tt.fields.repo,
				log:              tt.fields.log,
				salesforceClient: tt.fields.salesforceClient,
			}
			userMock := &entities.User{
				Email:                    tt.args.req.Email,
				FirstName:                tt.args.req.FirstName,
				LastName:                 tt.args.req.LastName,
				IAMUID:                   tt.args.req.IAMId,
				IAMProvider:              tt.args.req.IAM,
				Status:                   "new",
				HasSwipedRecommendations: false,
			}
			mockRepo.EXPECT().Create(tt.args.ctx, userMock).Return(&entities.User{
				UserUUID:                 userId,
				FirstName:                tt.args.req.FirstName,
				LastName:                 tt.args.req.LastName,
				Email:                    tt.args.req.Email,
				IAMUID:                   tt.args.req.IAMId,
				IAMProvider:              entities.IAMProvider(tt.args.req.IAM),
				Status:                   "new",
				HasSwipedRecommendations: false,
			}, nil)
			mockSFClient.EXPECT().CreateUserAccount(gomock.Any(), &sfEntities.CreateSFUserRequest{
				PersonEmail: tt.args.req.Email,
				FirstName:   tt.args.req.FirstName,
				LastName:    *tt.args.req.LastName,
			}).Return(&sfEntities.CreateSFUserResponse{
				ID: "demo-sf-user-id",
			}, nil).AnyTimes()
			mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			got, err := s.CreateUser(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.CreateUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_UpdateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUUID := uuid.New()
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()
	phone := gofakeit.Phone()
	profile := gofakeit.Word()
	sfId := "demo-sf-user-id"
	ctx := meta.WithXCustomerID(context.Background(), userUUID.String())

	setup := func() (*service, context.Context, *repository.MockRepository, *salesforce.MockClient) {
		mockRepo := repository.NewMockRepository(ctrl)
		mockSFClient := salesforce.NewMockClient(ctrl)
		svc := &service{
			repo:             mockRepo,
			log:              zap.NewExample().Sugar(),
			config:           cfg.Config{CDNDomain: "cdn.bookstore.ai"},
			salesforceClient: mockSFClient,
		}
		return svc, ctx, mockRepo, mockSFClient
	}

	t.Run("user not found in request", func(t *testing.T) {
		svc, _, _, _ := setup()
		_, err := svc.UpdateUser(context.Background(), &entities.UpdateUserRequest{
			Data: &entities.UpdateUserRequestBody{},
		})
		assert.Error(t, err)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		svc, ctx, mockRepo, mockSFClient := setup()
		req := &entities.UpdateUserRequest{
			Data: &entities.UpdateUserRequestBody{
				FirstName:  firstName,
				LastName:   &lastName,
				Phone:      &phone,
				ProfilePic: &profile,
			},
		}
		expectedUser := &entities.User{
			UserUUID:   userUUID,
			FirstName:  firstName,
			LastName:   &lastName,
			Phone:      phone,
			ProfilePic: profile,
			SFUserID:   &sfId,
		}

		mockRepo.EXPECT().Update(ctx, map[string]interface{}{
			"first_name":  firstName,
			"last_name":   &lastName,
			"phone":       &phone,
			"profile_pic": &profile,
		}, userUUID.String()).Return(nil)
		mockRepo.EXPECT().FindByUUID(ctx, userUUID.String()).Return(expectedUser, nil)
		mockSFClient.EXPECT().UpdateUserAccount(gomock.Any(), &sfEntities.UpdateSFUserRequest{
			ID:        sfId,
			FirstName: req.Data.FirstName,
			LastName:  *req.Data.LastName,
			Phone:     *req.Data.Phone,
		}).Return(nil).AnyTimes()

		got, err := svc.UpdateUser(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, got)
	})

	t.Run("UpdateUserWithOnlyFirstName", func(t *testing.T) {
		svc, ctx, mockRepo, mockSFClient := setup()
		req := &entities.UpdateUserRequest{
			Data: &entities.UpdateUserRequestBody{
				FirstName:  firstName,
				LastName:   nil,
				Phone:      nil,
				ProfilePic: nil,
			},
		}
		expectedUser := &entities.User{
			UserUUID:   userUUID,
			FirstName:  firstName,
			LastName:   nil,
			Phone:      "",
			ProfilePic: "",
			SFUserID:   &sfId,
		}

		mockRepo.EXPECT().Update(ctx, map[string]interface{}{
			"first_name":  req.Data.FirstName,
			"last_name":   req.Data.LastName,
			"phone":       req.Data.Phone,
			"profile_pic": req.Data.ProfilePic,
		}, userUUID.String()).Return(nil)
		mockRepo.EXPECT().FindByUUID(ctx, userUUID.String()).Return(expectedUser, nil)
		mockSFClient.EXPECT().UpdateUserAccount(gomock.Any(), &sfEntities.UpdateSFUserRequest{
			ID:        sfId,
			FirstName: req.Data.FirstName,
			LastName:  "\u200b",
		}).Return(nil).AnyTimes()

		got, err := svc.UpdateUser(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, got)
	})

	t.Run("UpdateUserProfilePic", func(t *testing.T) {
		svc, ctx, mockRepo, mockSFClient := setup()
		req := &entities.UpdateUserRequest{
			Data: &entities.UpdateUserRequestBody{
				FirstName:  firstName,
				LastName:   nil,
				Phone:      nil,
				ProfilePic: &profile,
			},
		}
		expectedUser := &entities.User{
			UserUUID:   userUUID,
			FirstName:  firstName,
			LastName:   nil,
			Phone:      "",
			ProfilePic: fmt.Sprintf("https://cdn.bookstore.ai/%s", profile),
			SFUserID:   &sfId,
		}

		mockRepo.EXPECT().Update(ctx, map[string]interface{}{
			"first_name":  req.Data.FirstName,
			"last_name":   req.Data.LastName,
			"phone":       req.Data.Phone,
			"profile_pic": req.Data.ProfilePic,
		}, userUUID.String()).Return(nil)
		mockRepo.EXPECT().FindByUUID(ctx, userUUID.String()).Return(expectedUser, nil)
		mockSFClient.EXPECT().UpdateUserAccount(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		got, err := svc.UpdateUser(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, got)
	})

	t.Run("repository update error", func(t *testing.T) {
		svc, ctx, mockRepo, _ := setup()
		req := &entities.UpdateUserRequest{
			Data: &entities.UpdateUserRequestBody{
				FirstName:  firstName,
				LastName:   &lastName,
				Phone:      &phone,
				ProfilePic: &profile,
			},
		}

		mockRepo.EXPECT().Update(ctx, map[string]interface{}{
			"first_name":  firstName,
			"last_name":   &lastName,
			"phone":       &phone,
			"profile_pic": &profile,
		}, userUUID.String()).Return(errors.New("repository error"))

		_, err := svc.UpdateUser(ctx, req)
		assert.Error(t, err)
	})

	t.Run("repository find error", func(t *testing.T) {
		svc, ctx, mockRepo, _ := setup()
		req := &entities.UpdateUserRequest{
			Data: &entities.UpdateUserRequestBody{
				FirstName:  firstName,
				LastName:   &lastName,
				Phone:      &phone,
				ProfilePic: &profile,
			},
		}

		mockRepo.EXPECT().Update(ctx, map[string]interface{}{
			"first_name":  firstName,
			"last_name":   &lastName,
			"phone":       &phone,
			"profile_pic": &profile,
		}, userUUID.String()).Return(nil)
		mockRepo.EXPECT().FindByUUID(ctx, userUUID.String()).Return(nil, errors.New("repository error"))

		_, err := svc.UpdateUser(ctx, req)
		assert.Error(t, err)
	})
}

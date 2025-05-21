package service

import (
	"context"
	authorizenetConfig "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/config"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/stripe/entities"
	"go.uber.org/zap"
)

type Service interface {
	CreateCustomer(_ context.Context, req *entities.CreateCustomerRequest) (*entities.CreateCustomerResponse, error)
	GetCustomerPaymentMethods(_ context.Context, customerId *string) (*entities.GetCustomerPaymentMethodsResponse, error)
	GetSetupIntent(_ context.Context, customerId *string) (*entities.GetSetupIntentResponse, error)
	CreatePaymentIntent(ctx context.Context, req *entities.CreatePaymentIntentRequest) (*entities.CreatePaymentIntentResponse, error)
	GetWebhookEvent(_ context.Context, req *entities.HandleWebhookEventRequest) (*entities.HandleWebhookEventResponse, error)
}

func New(config authorizenetConfig.Config, logger *zap.SugaredLogger) (Service, error) {
	return &service{config, logger}, nil
}

type service struct {
	config authorizenetConfig.Config
	logger *zap.SugaredLogger
}

func (s *service) CreateCustomer(_ context.Context, req *entities.CreateCustomerRequest) (*entities.CreateCustomerResponse, error) {
	return nil, nil
}

func (s *service) GetCustomerPaymentMethods(_ context.Context, customerId *string) (*entities.GetCustomerPaymentMethodsResponse, error) {
	return nil, nil
}

func (s *service) GetSetupIntent(_ context.Context, customerId *string) (*entities.GetSetupIntentResponse, error) {
	return nil, nil
}

func (s *service) CreatePaymentIntent(_ context.Context, req *entities.CreatePaymentIntentRequest) (*entities.CreatePaymentIntentResponse, error) {
	return nil, nil
}

func (s *service) GetWebhookEvent(_ context.Context, req *entities.HandleWebhookEventRequest) (*entities.HandleWebhookEventResponse, error) {
	return nil, nil
}

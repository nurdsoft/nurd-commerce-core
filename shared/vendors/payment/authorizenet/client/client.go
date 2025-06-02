package client

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/service"
	"github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/providers"
	"github.com/pkg/errors"
)

type Client interface {
	CreateCustomer(ctx context.Context, req entities.CreateCustomerRequest) (entities.CreateCustomerResponse, error)
	CreateCustomerPaymentProfile(ctx context.Context, req entities.CreateCustomerPaymentProfileRequest) (entities.CreateCustomerPaymentProfileResponse, error)
	GetCustomerPaymentMethods(ctx context.Context, req entities.GetPaymentProfilesRequest) (entities.GetPaymentProfilesResponse, error)
	CreatePayment(ctx context.Context, req any) (providers.PaymentProviderResponse, error)
	GetProvider() providers.ProviderType
}

func NewClient(svc service.Service) Client {
	return &localClient{svc}
}

type localClient struct {
	svc service.Service
}

func (c *localClient) CreateCustomer(ctx context.Context, req entities.CreateCustomerRequest) (entities.CreateCustomerResponse, error) {
	return c.svc.CreateCustomerProfile(ctx, req)
}

func (c *localClient) CreateCustomerPaymentProfile(ctx context.Context, req entities.CreateCustomerPaymentProfileRequest) (entities.CreateCustomerPaymentProfileResponse, error) {
	return c.svc.CreateCustomerPaymentProfile(ctx, req)
}

func (c *localClient) GetCustomerPaymentMethods(ctx context.Context, req entities.GetPaymentProfilesRequest) (entities.GetPaymentProfilesResponse, error) {
	return c.svc.GetCustomerPaymentProfiles(ctx, req)
}

func (c *localClient) CreatePayment(ctx context.Context, req any) (providers.PaymentProviderResponse, error) {
	authorizeNetReq, ok := req.(entities.CreatePaymentTransactionRequest)
	if !ok {
		return providers.PaymentProviderResponse{}, errors.New("invalid payment request type")
	}

	res, err := c.svc.CreatePaymentTransaction(ctx, authorizeNetReq)
	if err != nil {
		return providers.PaymentProviderResponse{}, err
	}

	return providers.PaymentProviderResponse{
		ID:     res.ID,
		Status: mapAuthorizeNetStatusToPaymentStatus(res.Status),
	}, nil
}

func mapAuthorizeNetStatusToPaymentStatus(status string) providers.PaymentStatus {
	switch status {
	case service.AuthorizeNetStatusApproved:
		return providers.PaymentStatusSuccess
	case service.AuthorizeNetStatusDeclined, service.AuthorizeNetStatusError:
		return providers.PaymentStatusFailed
	case service.AuthorizeNetStatusHeldForReview:
		return providers.PaymentStatusPending
	}

	return providers.PaymentStatusFailed
}

func (c *localClient) GetProvider() providers.ProviderType {
	return providers.ProviderAuthorizeNet
}

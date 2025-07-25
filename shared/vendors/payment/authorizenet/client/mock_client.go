// Code generated by MockGen. DO NOT EDIT.
// Source: shared/vendors/payment/authorizenet/client/client.go

// Package client is a generated GoMock package.
package client

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	entities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/authorizenet/entities"
	providers "github.com/nurdsoft/nurd-commerce-core/shared/vendors/payment/providers"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// CreateCustomer mocks base method.
func (m *MockClient) CreateCustomer(ctx context.Context, req entities.CreateCustomerRequest) (entities.CreateCustomerResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCustomer", ctx, req)
	ret0, _ := ret[0].(entities.CreateCustomerResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateCustomer indicates an expected call of CreateCustomer.
func (mr *MockClientMockRecorder) CreateCustomer(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCustomer", reflect.TypeOf((*MockClient)(nil).CreateCustomer), ctx, req)
}

// CreateCustomerPaymentProfile mocks base method.
func (m *MockClient) CreateCustomerPaymentProfile(ctx context.Context, req entities.CreateCustomerPaymentProfileRequest) (entities.CreateCustomerPaymentProfileResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCustomerPaymentProfile", ctx, req)
	ret0, _ := ret[0].(entities.CreateCustomerPaymentProfileResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateCustomerPaymentProfile indicates an expected call of CreateCustomerPaymentProfile.
func (mr *MockClientMockRecorder) CreateCustomerPaymentProfile(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCustomerPaymentProfile", reflect.TypeOf((*MockClient)(nil).CreateCustomerPaymentProfile), ctx, req)
}

// CreatePayment mocks base method.
func (m *MockClient) CreatePayment(ctx context.Context, req any) (providers.PaymentProviderResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePayment", ctx, req)
	ret0, _ := ret[0].(providers.PaymentProviderResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreatePayment indicates an expected call of CreatePayment.
func (mr *MockClientMockRecorder) CreatePayment(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePayment", reflect.TypeOf((*MockClient)(nil).CreatePayment), ctx, req)
}

// GetCustomerPaymentMethods mocks base method.
func (m *MockClient) GetCustomerPaymentMethods(ctx context.Context, req entities.GetPaymentProfilesRequest) (entities.GetPaymentProfilesResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCustomerPaymentMethods", ctx, req)
	ret0, _ := ret[0].(entities.GetPaymentProfilesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCustomerPaymentMethods indicates an expected call of GetCustomerPaymentMethods.
func (mr *MockClientMockRecorder) GetCustomerPaymentMethods(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCustomerPaymentMethods", reflect.TypeOf((*MockClient)(nil).GetCustomerPaymentMethods), ctx, req)
}

// GetProvider mocks base method.
func (m *MockClient) GetProvider() providers.ProviderType {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProvider")
	ret0, _ := ret[0].(providers.ProviderType)
	return ret0
}

// GetProvider indicates an expected call of GetProvider.
func (mr *MockClientMockRecorder) GetProvider() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProvider", reflect.TypeOf((*MockClient)(nil).GetProvider))
}

// Refund mocks base method.
func (m *MockClient) Refund(ctx context.Context, req any) (*providers.RefundResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Refund", ctx, req)
	ret0, _ := ret[0].(*providers.RefundResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Refund indicates an expected call of Refund.
func (mr *MockClientMockRecorder) Refund(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Refund", reflect.TypeOf((*MockClient)(nil).Refund), ctx, req)
}

// Code generated by MockGen. DO NOT EDIT.
// Source: internal/orders/ordersclient/client.go

// Package ordersclient is a generated GoMock package.
package ordersclient

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	entities "github.com/nurdsoft/nurd-commerce-core/internal/orders/entities"
	decimal "github.com/shopspring/decimal"
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

// ProcessOrderStatus mocks base method.
func (m *MockClient) ProcessOrderStatus(ctx context.Context, req *entities.UpdateOrderRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessOrderStatus", ctx, req)
	ret0, _ := ret[0].(error)
	return ret0
}

// ProcessOrderStatus indicates an expected call of ProcessOrderStatus.
func (mr *MockClientMockRecorder) ProcessOrderStatus(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessOrderStatus", reflect.TypeOf((*MockClient)(nil).ProcessOrderStatus), ctx, req)
}

// ProcessPaymentFailed mocks base method.
func (m *MockClient) ProcessPaymentFailed(ctx context.Context, paymentID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessPaymentFailed", ctx, paymentID)
	ret0, _ := ret[0].(error)
	return ret0
}

// ProcessPaymentFailed indicates an expected call of ProcessPaymentFailed.
func (mr *MockClientMockRecorder) ProcessPaymentFailed(ctx, paymentID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessPaymentFailed", reflect.TypeOf((*MockClient)(nil).ProcessPaymentFailed), ctx, paymentID)
}

// ProcessPaymentSucceeded mocks base method.
func (m *MockClient) ProcessPaymentSucceeded(ctx context.Context, paymentID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessPaymentSucceeded", ctx, paymentID)
	ret0, _ := ret[0].(error)
	return ret0
}

// ProcessPaymentSucceeded indicates an expected call of ProcessPaymentSucceeded.
func (mr *MockClientMockRecorder) ProcessPaymentSucceeded(ctx, paymentID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessPaymentSucceeded", reflect.TypeOf((*MockClient)(nil).ProcessPaymentSucceeded), ctx, paymentID)
}

// ProcessRefundSucceeded mocks base method.
func (m *MockClient) ProcessRefundSucceeded(ctx context.Context, refundId string, refundAmount decimal.Decimal) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessRefundSucceeded", ctx, refundId, refundAmount)
	ret0, _ := ret[0].(error)
	return ret0
}

// ProcessRefundSucceeded indicates an expected call of ProcessRefundSucceeded.
func (mr *MockClientMockRecorder) ProcessRefundSucceeded(ctx, refundId, refundAmount interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessRefundSucceeded", reflect.TypeOf((*MockClient)(nil).ProcessRefundSucceeded), ctx, refundId, refundAmount)
}

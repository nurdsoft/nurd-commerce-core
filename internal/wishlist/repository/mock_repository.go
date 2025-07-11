// Code generated by MockGen. DO NOT EDIT.
// Source: internal/wishlist/repository/repository.go

// Package repository is a generated GoMock package.
package repository

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
	entities "github.com/nurdsoft/nurd-commerce-core/internal/wishlist/entities"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// BulkRemoveFromWishlist mocks base method.
func (m *MockRepository) BulkRemoveFromWishlist(ctx context.Context, customerID uuid.UUID, productIDs []uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BulkRemoveFromWishlist", ctx, customerID, productIDs)
	ret0, _ := ret[0].(error)
	return ret0
}

// BulkRemoveFromWishlist indicates an expected call of BulkRemoveFromWishlist.
func (mr *MockRepositoryMockRecorder) BulkRemoveFromWishlist(ctx, customerID, productIDs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BulkRemoveFromWishlist", reflect.TypeOf((*MockRepository)(nil).BulkRemoveFromWishlist), ctx, customerID, productIDs)
}

// DeleteFromWishlist mocks base method.
func (m *MockRepository) DeleteFromWishlist(ctx context.Context, customerID string, productID uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteFromWishlist", ctx, customerID, productID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteFromWishlist indicates an expected call of DeleteFromWishlist.
func (mr *MockRepositoryMockRecorder) DeleteFromWishlist(ctx, customerID, productID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteFromWishlist", reflect.TypeOf((*MockRepository)(nil).DeleteFromWishlist), ctx, customerID, productID)
}

// GetWishlist mocks base method.
func (m *MockRepository) GetWishlist(ctx context.Context, customerID string, limit int, cursor string) ([]*entities.WishlistItem, string, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWishlist", ctx, customerID, limit, cursor)
	ret0, _ := ret[0].([]*entities.WishlistItem)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(int64)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// GetWishlist indicates an expected call of GetWishlist.
func (mr *MockRepositoryMockRecorder) GetWishlist(ctx, customerID, limit, cursor interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWishlist", reflect.TypeOf((*MockRepository)(nil).GetWishlist), ctx, customerID, limit, cursor)
}

// GetWishlistProductTimestamps mocks base method.
func (m *MockRepository) GetWishlistProductTimestamps(customerID string, productIDs []uuid.UUID) (map[string]time.Time, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWishlistProductTimestamps", customerID, productIDs)
	ret0, _ := ret[0].(map[string]time.Time)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWishlistProductTimestamps indicates an expected call of GetWishlistProductTimestamps.
func (mr *MockRepositoryMockRecorder) GetWishlistProductTimestamps(customerID, productIDs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWishlistProductTimestamps", reflect.TypeOf((*MockRepository)(nil).GetWishlistProductTimestamps), customerID, productIDs)
}

// UpdateWishlist mocks base method.
func (m *MockRepository) UpdateWishlist(ctx context.Context, customerID string, productIDs []uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateWishlist", ctx, customerID, productIDs)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateWishlist indicates an expected call of UpdateWishlist.
func (mr *MockRepositoryMockRecorder) UpdateWishlist(ctx, customerID, productIDs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateWishlist", reflect.TypeOf((*MockRepository)(nil).UpdateWishlist), ctx, customerID, productIDs)
}

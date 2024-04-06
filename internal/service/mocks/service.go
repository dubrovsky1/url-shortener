// Code generated by MockGen. DO NOT EDIT.
// Source: service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/dubrovsky1/url-shortener/internal/models"
	gomock "github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
)

// MockStorager is a mock of Storager interface.
type MockStorager struct {
	ctrl     *gomock.Controller
	recorder *MockStoragerMockRecorder
}

// MockStoragerMockRecorder is the mock recorder for MockStorager.
type MockStoragerMockRecorder struct {
	mock *MockStorager
}

// NewMockStorager creates a new mock instance.
func NewMockStorager(ctrl *gomock.Controller) *MockStorager {
	mock := &MockStorager{ctrl: ctrl}
	mock.recorder = &MockStoragerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorager) EXPECT() *MockStoragerMockRecorder {
	return m.recorder
}

// DeleteURL mocks base method.
func (m *MockStorager) DeleteURL(arg0 context.Context, arg1 []models.DeletedURLS) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteURL", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteURL indicates an expected call of DeleteURL.
func (mr *MockStoragerMockRecorder) DeleteURL(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteURL", reflect.TypeOf((*MockStorager)(nil).DeleteURL), arg0, arg1)
}

// GetURL mocks base method.
func (m *MockStorager) GetURL(arg0 context.Context, arg1 models.ShortURL) (models.ShortenURL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetURL", arg0, arg1)
	ret0, _ := ret[0].(models.ShortenURL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetURL indicates an expected call of GetURL.
func (mr *MockStoragerMockRecorder) GetURL(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetURL", reflect.TypeOf((*MockStorager)(nil).GetURL), arg0, arg1)
}

// InsertBatch mocks base method.
func (m *MockStorager) InsertBatch(arg0 context.Context, arg1 []models.BatchRequest, arg2 models.Host, arg3 uuid.UUID) ([]models.BatchResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertBatch", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]models.BatchResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertBatch indicates an expected call of InsertBatch.
func (mr *MockStoragerMockRecorder) InsertBatch(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertBatch", reflect.TypeOf((*MockStorager)(nil).InsertBatch), arg0, arg1, arg2, arg3)
}

// ListByUserID mocks base method.
func (m *MockStorager) ListByUserID(arg0 context.Context, arg1 models.Host, arg2 uuid.UUID) ([]models.ShortenURL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListByUserID", arg0, arg1, arg2)
	ret0, _ := ret[0].([]models.ShortenURL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListByUserID indicates an expected call of ListByUserID.
func (mr *MockStoragerMockRecorder) ListByUserID(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByUserID", reflect.TypeOf((*MockStorager)(nil).ListByUserID), arg0, arg1, arg2)
}

// SaveURL mocks base method.
func (m *MockStorager) SaveURL(arg0 context.Context, arg1 models.ShortenURL) (models.ShortURL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveURL", arg0, arg1)
	ret0, _ := ret[0].(models.ShortURL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveURL indicates an expected call of SaveURL.
func (mr *MockStoragerMockRecorder) SaveURL(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveURL", reflect.TypeOf((*MockStorager)(nil).SaveURL), arg0, arg1)
}

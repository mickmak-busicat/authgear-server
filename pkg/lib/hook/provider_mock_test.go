// Code generated by MockGen. DO NOT EDIT.
// Source: provider.go

// Package hook is a generated GoMock package.
package hook

import (
	event "github.com/authgear/authgear-server/pkg/api/event"
	model "github.com/authgear/authgear-server/pkg/api/model"
	db "github.com/authgear/authgear-server/pkg/lib/infra/db"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockUserProvider is a mock of UserProvider interface
type MockUserProvider struct {
	ctrl     *gomock.Controller
	recorder *MockUserProviderMockRecorder
}

// MockUserProviderMockRecorder is the mock recorder for MockUserProvider
type MockUserProviderMockRecorder struct {
	mock *MockUserProvider
}

// NewMockUserProvider creates a new mock instance
func NewMockUserProvider(ctrl *gomock.Controller) *MockUserProvider {
	mock := &MockUserProvider{ctrl: ctrl}
	mock.recorder = &MockUserProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockUserProvider) EXPECT() *MockUserProviderMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockUserProvider) Get(id string) (*model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", id)
	ret0, _ := ret[0].(*model.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockUserProviderMockRecorder) Get(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockUserProvider)(nil).Get), id)
}

// MockDeliverer is a mock of deliverer interface
type MockDeliverer struct {
	ctrl     *gomock.Controller
	recorder *MockDelivererMockRecorder
}

// MockDelivererMockRecorder is the mock recorder for MockDeliverer
type MockDelivererMockRecorder struct {
	mock *MockDeliverer
}

// NewMockDeliverer creates a new mock instance
func NewMockDeliverer(ctrl *gomock.Controller) *MockDeliverer {
	mock := &MockDeliverer{ctrl: ctrl}
	mock.recorder = &MockDelivererMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDeliverer) EXPECT() *MockDelivererMockRecorder {
	return m.recorder
}

// WillDeliver mocks base method
func (m *MockDeliverer) WillDeliver(eventType event.Type) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WillDeliver", eventType)
	ret0, _ := ret[0].(bool)
	return ret0
}

// WillDeliver indicates an expected call of WillDeliver
func (mr *MockDelivererMockRecorder) WillDeliver(eventType interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WillDeliver", reflect.TypeOf((*MockDeliverer)(nil).WillDeliver), eventType)
}

// DeliverBeforeEvent mocks base method
func (m *MockDeliverer) DeliverBeforeEvent(event *event.Event) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeliverBeforeEvent", event)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeliverBeforeEvent indicates an expected call of DeliverBeforeEvent
func (mr *MockDelivererMockRecorder) DeliverBeforeEvent(event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeliverBeforeEvent", reflect.TypeOf((*MockDeliverer)(nil).DeliverBeforeEvent), event)
}

// DeliverNonBeforeEvent mocks base method
func (m *MockDeliverer) DeliverNonBeforeEvent(event *event.Event) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeliverNonBeforeEvent", event)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeliverNonBeforeEvent indicates an expected call of DeliverNonBeforeEvent
func (mr *MockDelivererMockRecorder) DeliverNonBeforeEvent(event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeliverNonBeforeEvent", reflect.TypeOf((*MockDeliverer)(nil).DeliverNonBeforeEvent), event)
}

// MockStore is a mock of store interface
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
}

// MockStoreMockRecorder is the mock recorder for MockStore
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// NextSequenceNumber mocks base method
func (m *MockStore) NextSequenceNumber() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NextSequenceNumber")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NextSequenceNumber indicates an expected call of NextSequenceNumber
func (mr *MockStoreMockRecorder) NextSequenceNumber() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NextSequenceNumber", reflect.TypeOf((*MockStore)(nil).NextSequenceNumber))
}

// AddEvents mocks base method
func (m *MockStore) AddEvents(events []*event.Event) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddEvents", events)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddEvents indicates an expected call of AddEvents
func (mr *MockStoreMockRecorder) AddEvents(events interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddEvents", reflect.TypeOf((*MockStore)(nil).AddEvents), events)
}

// GetEventsForDelivery mocks base method
func (m *MockStore) GetEventsForDelivery() ([]*event.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEventsForDelivery")
	ret0, _ := ret[0].([]*event.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEventsForDelivery indicates an expected call of GetEventsForDelivery
func (mr *MockStoreMockRecorder) GetEventsForDelivery() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEventsForDelivery", reflect.TypeOf((*MockStore)(nil).GetEventsForDelivery))
}

// MockDatabaseHandle is a mock of DatabaseHandle interface
type MockDatabaseHandle struct {
	ctrl     *gomock.Controller
	recorder *MockDatabaseHandleMockRecorder
}

// MockDatabaseHandleMockRecorder is the mock recorder for MockDatabaseHandle
type MockDatabaseHandleMockRecorder struct {
	mock *MockDatabaseHandle
}

// NewMockDatabaseHandle creates a new mock instance
func NewMockDatabaseHandle(ctrl *gomock.Controller) *MockDatabaseHandle {
	mock := &MockDatabaseHandle{ctrl: ctrl}
	mock.recorder = &MockDatabaseHandleMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDatabaseHandle) EXPECT() *MockDatabaseHandleMockRecorder {
	return m.recorder
}

// UseHook mocks base method
func (m *MockDatabaseHandle) UseHook(hook db.TransactionHook) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "UseHook", hook)
}

// UseHook indicates an expected call of UseHook
func (mr *MockDatabaseHandleMockRecorder) UseHook(hook interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UseHook", reflect.TypeOf((*MockDatabaseHandle)(nil).UseHook), hook)
}

// Code generated by MockGen. DO NOT EDIT.
// Source: k8s.io/client-go/rest (interfaces: ResponseWrapper)

// Package debugutils is a generated GoMock package.
package debugutils

import (
	context "context"
	io "io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockResponseWrapper is a mock of ResponseWrapper interface
type MockResponseWrapper struct {
	ctrl     *gomock.Controller
	recorder *MockResponseWrapperMockRecorder
}

// MockResponseWrapperMockRecorder is the mock recorder for MockResponseWrapper
type MockResponseWrapperMockRecorder struct {
	mock *MockResponseWrapper
}

// NewMockResponseWrapper creates a new mock instance
func NewMockResponseWrapper(ctrl *gomock.Controller) *MockResponseWrapper {
	mock := &MockResponseWrapper{ctrl: ctrl}
	mock.recorder = &MockResponseWrapperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockResponseWrapper) EXPECT() *MockResponseWrapperMockRecorder {
	return m.recorder
}

// DoRaw mocks base method
func (m *MockResponseWrapper) DoRaw(arg0 context.Context) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DoRaw", arg0)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DoRaw indicates an expected call of DoRaw
func (mr *MockResponseWrapperMockRecorder) DoRaw(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DoRaw", reflect.TypeOf((*MockResponseWrapper)(nil).DoRaw), arg0)
}

// Stream mocks base method
func (m *MockResponseWrapper) Stream(arg0 context.Context) (io.ReadCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stream", arg0)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Stream indicates an expected call of Stream
func (mr *MockResponseWrapperMockRecorder) Stream(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stream", reflect.TypeOf((*MockResponseWrapper)(nil).Stream), arg0)
}

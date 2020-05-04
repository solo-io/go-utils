// Code generated by MockGen. DO NOT EDIT.
// Source: ./helm_loaders.go

// Package mock_internal is a generated GoMock package.
package mock_internal

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "github.com/solo-io/go-utils/installutils/helminstall/types"
	action "helm.sh/helm/v3/pkg/action"
	chart "helm.sh/helm/v3/pkg/chart"
	cli "helm.sh/helm/v3/pkg/cli"
)

// MockActionConfigFactory is a mock of ActionConfigFactory interface
type MockActionConfigFactory struct {
	ctrl     *gomock.Controller
	recorder *MockActionConfigFactoryMockRecorder
}

// MockActionConfigFactoryMockRecorder is the mock recorder for MockActionConfigFactory
type MockActionConfigFactoryMockRecorder struct {
	mock *MockActionConfigFactory
}

// NewMockActionConfigFactory creates a new mock instance
func NewMockActionConfigFactory(ctrl *gomock.Controller) *MockActionConfigFactory {
	mock := &MockActionConfigFactory{ctrl: ctrl}
	mock.recorder = &MockActionConfigFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockActionConfigFactory) EXPECT() *MockActionConfigFactoryMockRecorder {
	return m.recorder
}

// NewActionConfigFromFile mocks base method
func (m *MockActionConfigFactory) NewActionConfig(kubeConfig, helmKubeContext, namespace string) (*action.Configuration, *cli.EnvSettings, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewActionConfigFromFile", kubeConfig, helmKubeContext, namespace)
	ret0, _ := ret[0].(*action.Configuration)
	ret1, _ := ret[1].(*cli.EnvSettings)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// NewActionConfigFromFile indicates an expected call of NewActionConfigFromFile
func (mr *MockActionConfigFactoryMockRecorder) NewActionConfig(kubeConfig, helmKubeContext, namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewActionConfigFromFile", reflect.TypeOf((*MockActionConfigFactory)(nil).NewActionConfig), kubeConfig, helmKubeContext, namespace)
}

// MockActionListFactory is a mock of ActionListFactory interface
type MockActionListFactory struct {
	ctrl     *gomock.Controller
	recorder *MockActionListFactoryMockRecorder
}

// MockActionListFactoryMockRecorder is the mock recorder for MockActionListFactory
type MockActionListFactoryMockRecorder struct {
	mock *MockActionListFactory
}

// NewMockActionListFactory creates a new mock instance
func NewMockActionListFactory(ctrl *gomock.Controller) *MockActionListFactory {
	mock := &MockActionListFactory{ctrl: ctrl}
	mock.recorder = &MockActionListFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockActionListFactory) EXPECT() *MockActionListFactoryMockRecorder {
	return m.recorder
}

// ReleaseList mocks base method
func (m *MockActionListFactory) ReleaseList(kubeConfig, helmKubeContext, namespace string) (types.ReleaseListRunner, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReleaseList", kubeConfig, helmKubeContext, namespace)
	ret0, _ := ret[0].(types.ReleaseListRunner)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReleaseList indicates an expected call of ReleaseList
func (mr *MockActionListFactoryMockRecorder) ReleaseList(kubeConfig, helmKubeContext, namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReleaseList", reflect.TypeOf((*MockActionListFactory)(nil).ReleaseList), kubeConfig, helmKubeContext, namespace)
}

// MockChartLoader is a mock of ChartLoader interface
type MockChartLoader struct {
	ctrl     *gomock.Controller
	recorder *MockChartLoaderMockRecorder
}

// MockChartLoaderMockRecorder is the mock recorder for MockChartLoader
type MockChartLoaderMockRecorder struct {
	mock *MockChartLoader
}

// NewMockChartLoader creates a new mock instance
func NewMockChartLoader(ctrl *gomock.Controller) *MockChartLoader {
	mock := &MockChartLoader{ctrl: ctrl}
	mock.recorder = &MockChartLoaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockChartLoader) EXPECT() *MockChartLoaderMockRecorder {
	return m.recorder
}

// Load mocks base method
func (m *MockChartLoader) Load(name string) (*chart.Chart, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Load", name)
	ret0, _ := ret[0].(*chart.Chart)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Load indicates an expected call of Load
func (mr *MockChartLoaderMockRecorder) Load(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Load", reflect.TypeOf((*MockChartLoader)(nil).Load), name)
}

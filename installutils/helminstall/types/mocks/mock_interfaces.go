// Code generated by MockGen. DO NOT EDIT.
// Source: ./interfaces.go

// Package mock_types is a generated GoMock package.
package mock_types

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "github.com/solo-io/go-utils/installutils/helminstall/types"
	chart "helm.sh/helm/v3/pkg/chart"
	cli "helm.sh/helm/v3/pkg/cli"
	release "helm.sh/helm/v3/pkg/release"
)

// MockHelmClient is a mock of HelmClient interface
type MockHelmClient struct {
	ctrl     *gomock.Controller
	recorder *MockHelmClientMockRecorder
}

// MockHelmClientMockRecorder is the mock recorder for MockHelmClient
type MockHelmClientMockRecorder struct {
	mock *MockHelmClient
}

// NewMockHelmClient creates a new mock instance
func NewMockHelmClient(ctrl *gomock.Controller) *MockHelmClient {
	mock := &MockHelmClient{ctrl: ctrl}
	mock.recorder = &MockHelmClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHelmClient) EXPECT() *MockHelmClientMockRecorder {
	return m.recorder
}

// NewInstall mocks base method
func (m *MockHelmClient) NewInstall(kubeConfig, kubeContext, namespace, releaseName string, dryRun bool) (types.HelmInstaller, *cli.EnvSettings, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewInstall", kubeConfig, kubeContext, namespace, releaseName, dryRun)
	ret0, _ := ret[0].(types.HelmInstaller)
	ret1, _ := ret[1].(*cli.EnvSettings)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// NewInstall indicates an expected call of NewInstall
func (mr *MockHelmClientMockRecorder) NewInstall(kubeConfig, kubeContext, namespace, releaseName, dryRun interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewInstall", reflect.TypeOf((*MockHelmClient)(nil).NewInstall), kubeConfig, kubeContext, namespace, releaseName, dryRun)
}

// NewUninstall mocks base method
func (m *MockHelmClient) NewUninstall(kubeConfig, kubeContext, namespace string) (types.HelmUninstaller, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewUninstall", kubeConfig, kubeContext, namespace)
	ret0, _ := ret[0].(types.HelmUninstaller)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewUninstall indicates an expected call of NewUninstall
func (mr *MockHelmClientMockRecorder) NewUninstall(kubeConfig, kubeContext, namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewUninstall", reflect.TypeOf((*MockHelmClient)(nil).NewUninstall), kubeConfig, kubeContext, namespace)
}

// ReleaseList mocks base method
func (m *MockHelmClient) ReleaseList(kubeConfig, kubeContext, namespace string) (types.ReleaseListRunner, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReleaseList", kubeConfig, kubeContext, namespace)
	ret0, _ := ret[0].(types.ReleaseListRunner)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReleaseList indicates an expected call of ReleaseList
func (mr *MockHelmClientMockRecorder) ReleaseList(kubeConfig, kubeContext, namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReleaseList", reflect.TypeOf((*MockHelmClient)(nil).ReleaseList), kubeConfig, kubeContext, namespace)
}

// DownloadChart mocks base method
func (m *MockHelmClient) DownloadChart(chartArchiveUri string) (*chart.Chart, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadChart", chartArchiveUri)
	ret0, _ := ret[0].(*chart.Chart)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DownloadChart indicates an expected call of DownloadChart
func (mr *MockHelmClientMockRecorder) DownloadChart(chartArchiveUri interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadChart", reflect.TypeOf((*MockHelmClient)(nil).DownloadChart), chartArchiveUri)
}

// ReleaseExists mocks base method
func (m *MockHelmClient) ReleaseExists(kubeConfig, kubeContext, namespace, releaseName string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReleaseExists", kubeConfig, kubeContext, namespace, releaseName)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReleaseExists indicates an expected call of ReleaseExists
func (mr *MockHelmClientMockRecorder) ReleaseExists(kubeConfig, kubeContext, namespace, releaseName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReleaseExists", reflect.TypeOf((*MockHelmClient)(nil).ReleaseExists), kubeConfig, kubeContext, namespace, releaseName)
}

// MockHelmInstaller is a mock of HelmInstaller interface
type MockHelmInstaller struct {
	ctrl     *gomock.Controller
	recorder *MockHelmInstallerMockRecorder
}

// MockHelmInstallerMockRecorder is the mock recorder for MockHelmInstaller
type MockHelmInstallerMockRecorder struct {
	mock *MockHelmInstaller
}

// NewMockHelmInstaller creates a new mock instance
func NewMockHelmInstaller(ctrl *gomock.Controller) *MockHelmInstaller {
	mock := &MockHelmInstaller{ctrl: ctrl}
	mock.recorder = &MockHelmInstallerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHelmInstaller) EXPECT() *MockHelmInstallerMockRecorder {
	return m.recorder
}

// Run mocks base method
func (m *MockHelmInstaller) Run(chrt *chart.Chart, vals map[string]interface{}) (*release.Release, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", chrt, vals)
	ret0, _ := ret[0].(*release.Release)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Run indicates an expected call of Run
func (mr *MockHelmInstallerMockRecorder) Run(chrt, vals interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockHelmInstaller)(nil).Run), chrt, vals)
}

// MockHelmUninstaller is a mock of HelmUninstaller interface
type MockHelmUninstaller struct {
	ctrl     *gomock.Controller
	recorder *MockHelmUninstallerMockRecorder
}

// MockHelmUninstallerMockRecorder is the mock recorder for MockHelmUninstaller
type MockHelmUninstallerMockRecorder struct {
	mock *MockHelmUninstaller
}

// NewMockHelmUninstaller creates a new mock instance
func NewMockHelmUninstaller(ctrl *gomock.Controller) *MockHelmUninstaller {
	mock := &MockHelmUninstaller{ctrl: ctrl}
	mock.recorder = &MockHelmUninstallerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHelmUninstaller) EXPECT() *MockHelmUninstallerMockRecorder {
	return m.recorder
}

// Run mocks base method
func (m *MockHelmUninstaller) Run(name string) (*release.UninstallReleaseResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", name)
	ret0, _ := ret[0].(*release.UninstallReleaseResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Run indicates an expected call of Run
func (mr *MockHelmUninstallerMockRecorder) Run(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockHelmUninstaller)(nil).Run), name)
}

// MockReleaseListRunner is a mock of ReleaseListRunner interface
type MockReleaseListRunner struct {
	ctrl     *gomock.Controller
	recorder *MockReleaseListRunnerMockRecorder
}

// MockReleaseListRunnerMockRecorder is the mock recorder for MockReleaseListRunner
type MockReleaseListRunnerMockRecorder struct {
	mock *MockReleaseListRunner
}

// NewMockReleaseListRunner creates a new mock instance
func NewMockReleaseListRunner(ctrl *gomock.Controller) *MockReleaseListRunner {
	mock := &MockReleaseListRunner{ctrl: ctrl}
	mock.recorder = &MockReleaseListRunnerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockReleaseListRunner) EXPECT() *MockReleaseListRunnerMockRecorder {
	return m.recorder
}

// Run mocks base method
func (m *MockReleaseListRunner) Run() ([]*release.Release, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run")
	ret0, _ := ret[0].([]*release.Release)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Run indicates an expected call of Run
func (mr *MockReleaseListRunnerMockRecorder) Run() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockReleaseListRunner)(nil).Run))
}

// SetFilter mocks base method
func (m *MockReleaseListRunner) SetFilter(filter string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetFilter", filter)
}

// SetFilter indicates an expected call of SetFilter
func (mr *MockReleaseListRunnerMockRecorder) SetFilter(filter interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetFilter", reflect.TypeOf((*MockReleaseListRunner)(nil).SetFilter), filter)
}

// MockInstaller is a mock of Installer interface
type MockInstaller struct {
	ctrl     *gomock.Controller
	recorder *MockInstallerMockRecorder
}

// MockInstallerMockRecorder is the mock recorder for MockInstaller
type MockInstallerMockRecorder struct {
	mock *MockInstaller
}

// NewMockInstaller creates a new mock instance
func NewMockInstaller(ctrl *gomock.Controller) *MockInstaller {
	mock := &MockInstaller{ctrl: ctrl}
	mock.recorder = &MockInstallerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInstaller) EXPECT() *MockInstallerMockRecorder {
	return m.recorder
}

// Install mocks base method
func (m *MockInstaller) Install(installerConfig *types.InstallerConfig) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Install", installerConfig)
	ret0, _ := ret[0].(error)
	return ret0
}

// Install indicates an expected call of Install
func (mr *MockInstallerMockRecorder) Install(installerConfig interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Install", reflect.TypeOf((*MockInstaller)(nil).Install), installerConfig)
}

package helminstall_test

import (
	"bytes"
	"io/ioutil"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/helminstall"
	mock_helminstall "github.com/solo-io/go-utils/installutils/helminstall/mocks"
	mock_afero "github.com/solo-io/go-utils/testutils/mocks/afero"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

var _ = Describe("helm install client", func() {
	const (
		namespace = "test-namespace"
	)

	var (
		ctrl                       *gomock.Controller
		mockFile                   *mock_afero.MockFile
		mockFs                     *mock_helminstall.MockFs
		mockResourceFetcher        *mock_helminstall.MockResourceFetcher
		mockHelmActionConfigLoader *mock_helminstall.MockActionConfigLoader
		mockHelmActionListLoader   *mock_helminstall.MockActionListLoader
		mockHelmChartLoader        *mock_helminstall.MockChartLoader
		mockHelmLoaders            helminstall.HelmFactories
		mockHelmReleaseListRunner  *mock_helminstall.MockReleaseListRunner
		helmClient                 helminstall.HelmClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockFile = mock_afero.NewMockFile(ctrl)
		mockFs = mock_helminstall.NewMockFs(ctrl)
		mockResourceFetcher = mock_helminstall.NewMockResourceFetcher(ctrl)
		mockHelmActionConfigLoader = mock_helminstall.NewMockActionConfigLoader(ctrl)
		mockHelmChartLoader = mock_helminstall.NewMockChartLoader(ctrl)
		mockHelmActionListLoader = mock_helminstall.NewMockActionListLoader(ctrl)
		mockHelmReleaseListRunner = mock_helminstall.NewMockReleaseListRunner(ctrl)
		mockHelmLoaders = helminstall.HelmFactories{
			ActionConfigFactory: mockHelmActionConfigLoader,
			ActionListFactory:   mockHelmActionListLoader,
			ChartLoader:         mockHelmChartLoader,
		}
		helmClient = helminstall.NewDefaultHelmClient(
			mockFs,
			mockResourceFetcher,
			mockHelmLoaders)
	})

	It("should correctly configure Helm installation", func() {
		namespace := "namespace"
		releaseName := "releaseName"
		dryRun := true
		mockHelmActionConfigLoader.
			EXPECT().
			NewActionConfig(namespace).
			Return(&action.Configuration{}, nil, nil)
		install, _, err := helmClient.NewInstall(namespace, releaseName, dryRun)
		helmInstall := install.(*action.Install)
		Expect(err).ToNot(HaveOccurred())
		Expect(helmInstall.Namespace).To(Equal(namespace))
		Expect(helmInstall.ReleaseName).To(Equal(releaseName))
		Expect(helmInstall.DryRun).To(Equal(dryRun))
		Expect(helmInstall.ClientOnly).To(Equal(dryRun))
	})

	It("should download Helm chart", func() {
		chartUri := "chartUri.tgz"
		chartFileContents := "test chart file"
		chartFile := ioutil.NopCloser(bytes.NewBufferString(chartFileContents))
		chartTempFilePath := "/tmp/temp-filename"
		expectedChart := &chart.Chart{}
		mockResourceFetcher.
			EXPECT().
			GetResource(chartUri).
			Return(chartFile, nil)
		mockFs.
			EXPECT().
			NewTempFile("", helminstall.TempChartPrefix).
			Return(mockFile, nil)
		mockFile.
			EXPECT().
			Name().
			Return(chartTempFilePath)
		mockFs.
			EXPECT().
			WriteFile(chartTempFilePath, []byte(chartFileContents), helminstall.TempChartFilePermissions).
			Return(nil)
		mockHelmChartLoader.
			EXPECT().
			Load(chartTempFilePath).
			Return(expectedChart, nil)
		mockFs.
			EXPECT().
			RemoveAll(chartTempFilePath).
			Return(nil)
		chart, err := helmClient.DownloadChart(chartUri)
		Expect(err).ToNot(HaveOccurred())
		Expect(chart).To(BeIdenticalTo(expectedChart))
	})

	It("can properly set cli env settings with namespace", func() {
		settings := helminstall.NewCLISettings(namespace)
		Expect(settings.Namespace()).To(Equal(namespace))
	})

	It("should return true when release exists", func() {
		actionConfig := &action.Configuration{}
		namespace := "namespace"
		releaseName := "release-name"
		releases := []*release.Release{
			{Name: releaseName},
		}
		mockHelmActionConfigLoader.
			EXPECT().
			NewActionConfig(namespace).
			Return(actionConfig, nil, nil)
		mockHelmActionListLoader.
			EXPECT().
			ReleaseList(namespace).
			Return(mockHelmReleaseListRunner, nil)
		mockHelmReleaseListRunner.
			EXPECT().
			SetFilter(releaseName).
			Return()
		mockHelmReleaseListRunner.
			EXPECT().
			Run().
			Return(releases, nil)
		exists, err := helmClient.ReleaseExists(namespace, releaseName)
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(BeTrue())
	})

	It("should return false if release does not exist", func() {
		actionConfig := &action.Configuration{}
		namespace := "namespace"
		releaseName := "release-name"
		releases := []*release.Release{
			{Name: ""},
		}
		mockHelmActionConfigLoader.
			EXPECT().
			NewActionConfig(namespace).
			Return(actionConfig, nil, nil)
		mockHelmActionListLoader.
			EXPECT().
			ReleaseList(namespace).
			Return(mockHelmReleaseListRunner, nil)
		mockHelmReleaseListRunner.
			EXPECT().
			SetFilter(releaseName).
			Return()
		mockHelmReleaseListRunner.
			EXPECT().
			Run().
			Return(releases, nil)
		exists, err := helmClient.ReleaseExists(namespace, releaseName)
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(BeFalse())
	})
})

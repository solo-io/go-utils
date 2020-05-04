package internal_test

import (
	"bytes"
	"io/ioutil"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/helminstall/internal"
	mock_internal "github.com/solo-io/go-utils/installutils/helminstall/internal/mocks"
	"github.com/solo-io/go-utils/installutils/helminstall/types"
	mock_types "github.com/solo-io/go-utils/installutils/helminstall/types/mocks"
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
		ctrl                        *gomock.Controller
		mockFile                    *mock_afero.MockFile
		mockFs                      *mock_internal.MockFsHelper
		mockResourceFetcher         *mock_internal.MockResourceFetcher
		mockHelmActionConfigFactory *mock_internal.MockActionConfigFactory
		mockHelmActionListFactory   *mock_internal.MockActionListFactory
		mockHelmChartLoader         *mock_internal.MockChartLoader
		mockHelmLoaders             internal.HelmFactories
		mockHelmReleaseListRunner   *mock_types.MockReleaseListRunner
		helmClient                  types.HelmClient
		helmKubeConfig              = "path/to/kubeconfig"
		helmKubeContext             = "helm-kube-context"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockFile = mock_afero.NewMockFile(ctrl)
		mockFs = mock_internal.NewMockFsHelper(ctrl)
		mockResourceFetcher = mock_internal.NewMockResourceFetcher(ctrl)
		mockHelmActionConfigFactory = mock_internal.NewMockActionConfigFactory(ctrl)
		mockHelmChartLoader = mock_internal.NewMockChartLoader(ctrl)
		mockHelmActionListFactory = mock_internal.NewMockActionListFactory(ctrl)
		mockHelmReleaseListRunner = mock_types.NewMockReleaseListRunner(ctrl)
		mockHelmLoaders = internal.HelmFactories{
			ActionConfigFactory: mockHelmActionConfigFactory,
			ActionListFactory:   mockHelmActionListFactory,
			ChartLoader:         mockHelmChartLoader,
		}
		helmClient = internal.NewHelmClientForFileConfig(
			mockFs,
			mockResourceFetcher,
			mockHelmLoaders)
	})

	It("should correctly configure Helm installation", func() {
		namespace := "namespace"
		releaseName := "releaseName"
		dryRun := true
		mockHelmActionConfigFactory.
			EXPECT().
			NewActionConfig(helmKubeConfig, helmKubeContext, namespace).
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
			NewTempFile("", internal.TempChartPrefix).
			Return(mockFile, nil)
		mockFile.
			EXPECT().
			Name().
			Return(chartTempFilePath)
		mockFs.
			EXPECT().
			WriteFile(chartTempFilePath, []byte(chartFileContents), internal.TempChartFilePermissions).
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
		settings := internal.NewCLISettings(helmKubeConfig, helmKubeContext, namespace)
		Expect(settings.Namespace()).To(Equal(namespace))
	})

	It("should return true when release exists", func() {
		actionConfig := &action.Configuration{}
		namespace := "namespace"
		releaseName := "release-name"
		releases := []*release.Release{
			{Name: releaseName},
		}
		mockHelmActionConfigFactory.
			EXPECT().
			NewActionConfig(helmKubeConfig, helmKubeContext, namespace).
			Return(actionConfig, nil, nil)
		mockHelmActionListFactory.
			EXPECT().
			ReleaseList(helmKubeConfig, helmKubeContext, namespace).
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
		mockHelmActionConfigFactory.
			EXPECT().
			NewActionConfig(helmKubeConfig, helmKubeContext, namespace).
			Return(actionConfig, nil, nil)
		mockHelmActionListFactory.
			EXPECT().
			ReleaseList(helmKubeConfig, helmKubeContext, namespace).
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

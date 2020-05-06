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
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = Describe("helm install client", func() {
	const (
		namespace = "test-namespace"
	)

	var (
		ctrl                        *gomock.Controller
		mockResourceFetcher         *mock_internal.MockResourceFetcher
		mockHelmActionConfigFactory *mock_internal.MockActionConfigFactory
		mockHelmActionListFactory   *mock_internal.MockActionListFactory
		mockHelmChartLoader         *mock_internal.MockChartLoader
		mockHelmLoaders             internal.HelmFactories
		mockHelmReleaseListRunner   *mock_types.MockReleaseListRunner
		helmClientFromFileConfig    types.HelmClient
		helmClientFromMemoryConfig  types.HelmClient
		helmKubeConfigPath          = "path/to/kubeconfig"
		helmKubeContext             = "helm-kube-context"
		helmKubeConfig              = &clientcmd.DirectClientConfig{}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
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
		helmClientFromFileConfig = internal.NewHelmClientForFileConfig(
			mockResourceFetcher,
			mockHelmLoaders,
			helmKubeConfigPath,
			helmKubeContext,
		)
		helmClientFromMemoryConfig = internal.NewHelmClientForMemoryConfig(
			mockResourceFetcher,
			mockHelmLoaders,
			helmKubeConfig,
		)
	})

	It("should correctly configure Helm installation for file kubeconfig", func() {
		namespace := "namespace"
		releaseName := "releaseName"
		dryRun := true
		mockHelmActionConfigFactory.
			EXPECT().
			NewActionConfigFromFile(helmKubeConfigPath, helmKubeContext, namespace).
			Return(&action.Configuration{}, nil, nil)
		install, _, err := helmClientFromFileConfig.NewInstall(namespace, releaseName, dryRun)
		helmInstall := install.(*action.Install)
		Expect(err).ToNot(HaveOccurred())
		Expect(helmInstall.Namespace).To(Equal(namespace))
		Expect(helmInstall.ReleaseName).To(Equal(releaseName))
		Expect(helmInstall.DryRun).To(Equal(dryRun))
		Expect(helmInstall.ClientOnly).To(Equal(dryRun))
	})

	It("should correctly configure Helm installation for in memory kubeconfig", func() {
		namespace := "namespace"
		releaseName := "releaseName"
		dryRun := true
		mockHelmActionConfigFactory.
			EXPECT().
			NewActionConfigFromMemory(helmKubeConfig, namespace).
			Return(&action.Configuration{}, nil, nil)
		install, _, err := helmClientFromMemoryConfig.NewInstall(namespace, releaseName, dryRun)
		helmInstall := install.(*action.Install)
		Expect(err).ToNot(HaveOccurred())
		Expect(helmInstall.Namespace).To(Equal(namespace))
		Expect(helmInstall.ReleaseName).To(Equal(releaseName))
		Expect(helmInstall.DryRun).To(Equal(dryRun))
		Expect(helmInstall.ClientOnly).To(Equal(dryRun))
	})

	It("should correctly configure Helm un-installation", func() {
		mockHelmActionConfigFactory.
			EXPECT().
			NewActionConfigFromFile(helmKubeConfigPath, helmKubeContext, namespace).
			Return(&action.Configuration{}, nil, nil)
		_, err := helmClientFromFileConfig.NewUninstall(namespace)
		Expect(err).To(BeNil())
	})

	It("should download Helm chart", func() {
		chartUri := "chartUri.tgz"
		chartFileContents := "test chart file"
		chartFile := ioutil.NopCloser(bytes.NewBufferString(chartFileContents))
		expectedChart := &chart.Chart{}
		mockResourceFetcher.
			EXPECT().
			GetResource(chartUri).
			Return(chartFile, nil)
		mockHelmChartLoader.
			EXPECT().
			Load(chartFile).
			Return(expectedChart, nil)
		chart, err := helmClientFromFileConfig.DownloadChart(chartUri)
		Expect(err).ToNot(HaveOccurred())
		Expect(chart).To(BeIdenticalTo(expectedChart))
	})

	It("can properly set cli env settings with namespace", func() {
		settings := internal.NewCLISettings(helmKubeConfigPath, helmKubeContext, namespace)
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
			NewActionConfigFromFile(helmKubeConfigPath, helmKubeContext, namespace).
			Return(actionConfig, nil, nil)
		mockHelmActionListFactory.
			EXPECT().
			ReleaseList(actionConfig, namespace).
			Return(mockHelmReleaseListRunner)
		mockHelmReleaseListRunner.
			EXPECT().
			SetFilter(releaseName).
			Return()
		mockHelmReleaseListRunner.
			EXPECT().
			Run().
			Return(releases, nil)
		exists, err := helmClientFromFileConfig.ReleaseExists(namespace, releaseName)
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
			NewActionConfigFromFile(helmKubeConfigPath, helmKubeContext, namespace).
			Return(actionConfig, nil, nil)
		mockHelmActionListFactory.
			EXPECT().
			ReleaseList(actionConfig, namespace).
			Return(mockHelmReleaseListRunner)
		mockHelmReleaseListRunner.
			EXPECT().
			SetFilter(releaseName).
			Return()
		mockHelmReleaseListRunner.
			EXPECT().
			Run().
			Return(releases, nil)
		exists, err := helmClientFromFileConfig.ReleaseExists(namespace, releaseName)
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(BeFalse())
	})
})

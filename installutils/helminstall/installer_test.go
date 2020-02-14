package helminstall_test

import (
	"bytes"
	"os"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/helminstall"
	mock_helminstall "github.com/solo-io/go-utils/installutils/helminstall/mocks"
	"github.com/solo-io/go-utils/testutils"
)

var _ = Describe("Helm Installer", func() {
	var (
		ctrl                *gomock.Controller
		mockHelmClient      *mock_helminstall.MockHelmClient
		mockNamespaceClient *mock_helminstall.MockNamespaceCLient
		mockHelmInstaller   *mock_helminstall.MockHelmInstaller
		outputWriter        *bytes.Buffer
		installer           helminstall.Installer
		helmKubeconfig      = "path/to/kubeconfig"
		helmKubeContext     = "helm-kube-context"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockHelmClient = mock_helminstall.NewMockHelmClient(ctrl)
		mockNamespaceClient = mock_helminstall.NewMockNamespaceCLient(ctrl)
		mockHelmInstaller = mock_helminstall.NewMockHelmInstaller(ctrl)
		outputWriter = &bytes.Buffer{}
		installer = helminstall.NewInstaller(mockHelmClient, mockNamespaceClient, outputWriter)
	})

	It("should error if release already exists", func() {
		installerConfig := &helminstall.InstallerConfig{
			KubeConfig:       helmKubeconfig,
			KubeContext:      helmKubeContext,
			InstallNamespace: "namespace",
			ReleaseName:      "release-name",
			DryRun:           false,
		}
		mockHelmClient.
			EXPECT().
			ReleaseExists(helmKubeconfig, helmKubeContext, installerConfig.InstallNamespace, installerConfig.ReleaseName).
			Return(true, nil)
		err := installer.Install(installerConfig)
		Expect(err).To(testutils.HaveInErrorChain(
			helminstall.ReleaseAlreadyInstalledErr(installerConfig.ReleaseName, installerConfig.InstallNamespace)))
	})

	It("should install correctly", func() {
		installerConfig := &helminstall.InstallerConfig{
			KubeConfig:       helmKubeconfig,
			KubeContext:      helmKubeContext,
			InstallNamespace: "namespace",
			ReleaseName:      "release-name",
			ReleaseUri:       "release-uri",
			CreateNamespace:  true,
			DryRun:           false,
		}
		os.Setenv("HELM_NAMESPACE", "helm-namespace")
		defer os.Unsetenv("HELM_NAMESPACE")
		mockHelmClient.
			EXPECT().
			ReleaseExists(helmKubeconfig, helmKubeContext, installerConfig.InstallNamespace, installerConfig.ReleaseName).
			Return(false, nil)
		statusError := errors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}}
		mockNamespaceClient.
			EXPECT().
			Get(installerConfig.InstallNamespace, metav1.GetOptions{}).
			Return(nil, &statusError)
		mockNamespaceClient.
			EXPECT().
			Create(&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: installerConfig.InstallNamespace,
				},
			}).
			Return(nil, nil)
		mockHelmClient.
			EXPECT().
			NewInstall(helmKubeconfig, helmKubeContext, installerConfig.InstallNamespace, installerConfig.ReleaseName, installerConfig.DryRun).
			Return(mockHelmInstaller, cli.New(), nil)
		chartObj := &chart.Chart{}
		mockHelmClient.
			EXPECT().
			DownloadChart(installerConfig.ReleaseUri).
			Return(chartObj, nil)
		mockHelmInstaller.
			EXPECT().
			Run(chartObj, map[string]interface{}{}).
			Return(&release.Release{}, nil)

		err := installer.Install(installerConfig)
		Expect(err).NotTo(HaveOccurred())
	})
})

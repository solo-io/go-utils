package helminstall_test

import (
	"bytes"
	"context"
	"os"

	"github.com/solo-io/go-utils/installutils/helminstall/types"
	mock_types "github.com/solo-io/go-utils/installutils/helminstall/types/mocks"
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
		ctx                 context.Context
		ctrl                *gomock.Controller
		mockHelmClient      *mock_types.MockHelmClient
		mockNamespaceClient *mock_helminstall.MockNamespaceClient
		mockHelmInstaller   *mock_types.MockHelmInstaller
		outputWriter        *bytes.Buffer
		installer           types.Installer
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())
		mockHelmClient = mock_types.NewMockHelmClient(ctrl)
		mockNamespaceClient = mock_helminstall.NewMockNamespaceClient(ctrl)
		mockHelmInstaller = mock_types.NewMockHelmInstaller(ctrl)
		outputWriter = &bytes.Buffer{}
		installer = helminstall.NewInstaller(mockHelmClient, mockNamespaceClient, outputWriter)
	})

	It("should error if release already exists", func() {
		installerConfig := &types.InstallerConfig{
			InstallNamespace: "namespace",
			ReleaseName:      "release-name",
			DryRun:           false,
		}
		mockHelmClient.
			EXPECT().
			ReleaseExists(installerConfig.InstallNamespace, installerConfig.ReleaseName).
			Return(true, nil)
		err := installer.Install(ctx, installerConfig)
		Expect(err).To(testutils.HaveInErrorChain(
			helminstall.ReleaseAlreadyInstalledErr(installerConfig.ReleaseName, installerConfig.InstallNamespace)))
	})

	It("should install correctly", func() {
		installerConfig := &types.InstallerConfig{
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
			ReleaseExists(installerConfig.InstallNamespace, installerConfig.ReleaseName).
			Return(false, nil)
		statusError := errors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}}
		mockNamespaceClient.
			EXPECT().
			Get(ctx, installerConfig.InstallNamespace, metav1.GetOptions{}).
			Return(nil, &statusError)
		mockNamespaceClient.
			EXPECT().
			Create(ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: installerConfig.InstallNamespace,
				},
			}).
			Return(nil, nil)
		mockHelmClient.
			EXPECT().
			NewInstall(installerConfig.InstallNamespace, installerConfig.ReleaseName, installerConfig.DryRun).
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

		err := installer.Install(ctx, installerConfig)
		Expect(err).NotTo(HaveOccurred())
	})
})

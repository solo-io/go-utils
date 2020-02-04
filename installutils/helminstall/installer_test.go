package helminstall_test

import (
	"bytes"

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
		outputWriter        *bytes.Buffer
		installer           helminstall.Installer
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockHelmClient = mock_helminstall.NewMockHelmClient(ctrl)
		mockNamespaceClient = mock_helminstall.NewMockNamespaceCLient(ctrl)
		outputWriter = &bytes.Buffer{}
		installer = helminstall.NewInstaller(mockHelmClient, mockNamespaceClient, outputWriter)
	})

	It("should error if release already exists", func() {
		installerConfig := &helminstall.InstallerConfig{
			InstallNamespace: "namespace",
			ReleaseName:      "release-name",
			DryRun:           false,
		}
		mockHelmClient.
			EXPECT().
			ReleaseExists(installerConfig.InstallNamespace, installerConfig.ReleaseName).
			Return(true, nil)
		err := installer.Install(installerConfig)
		Expect(err).To(testutils.HaveInErrorChain(
			helminstall.ReleaseAlreadyInstalledErr(installerConfig.ReleaseName, installerConfig.InstallNamespace)))
	})
})

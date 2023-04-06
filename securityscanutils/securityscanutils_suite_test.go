package securityscanutils

import (
	"os/exec"
	"testing"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap/zapcore"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSecurityScanUtil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SecurityScanUtils Suite")
}

var _ = BeforeSuite(func() {
	// This test suite requires that Trivy is installed and present in the PATH
	contextutils.SetLogLevel(zapcore.DebugLevel)

	path, err := exec.LookPath("trivy")
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, path).NotTo(BeEmpty())
})

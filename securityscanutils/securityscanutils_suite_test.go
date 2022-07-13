package securityscanutils

import (
	"os/exec"
	"testing"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap/zapcore"

	"github.com/onsi/ginkgo/reporters"

	"github.com/solo-io/go-utils/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSecurityScanUtil(t *testing.T) {
	RegisterFailHandler(Fail)
	log.DefaultOut = GinkgoWriter
	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "SecurityScanUtils Suite", []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	// This test suite require that trivy is installed
	contextutils.SetLogLevel(zapcore.DebugLevel)

	path, err := exec.LookPath("trivy")
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, path).NotTo(BeEmpty())
})

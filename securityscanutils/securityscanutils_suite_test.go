package securityscanutils

import (
	"testing"

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

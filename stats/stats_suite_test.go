package stats_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/solo-io/go-utils/stats"

	"github.com/onsi/ginkgo/reporters"

	"github.com/solo-io/go-utils/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestStats(t *testing.T) {
	RegisterFailHandler(Fail)
	log.DefaultOut = GinkgoWriter
	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Stats Suite", []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	err := os.Setenv(stats.DefaultEnvVar, stats.DefaultEnabledValue)
	Expect(err).NotTo(HaveOccurred())

	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100
})

var _ = AfterSuite(func() {
	err := os.Unsetenv(stats.DefaultEnvVar)
	Expect(err).NotTo(HaveOccurred())
})

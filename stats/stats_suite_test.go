package stats_test

import (
	"os"
	"testing"

	"github.com/solo-io/go-utils/stats"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStats(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stats Suite")
}

var _ = BeforeSuite(func() {
	// The stats server only starts if appropriate environment variable is set
	err := os.Setenv(stats.DefaultEnvVar, stats.DefaultEnabledValue)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := os.Unsetenv(stats.DefaultEnvVar)
	Expect(err).NotTo(HaveOccurred())
})

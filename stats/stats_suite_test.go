package stats_test

import (
	"testing"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap/zapcore"

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
	// Tests in this suite expect the log level to be INFO to start
	contextutils.SetLogLevel(zapcore.InfoLevel)
})

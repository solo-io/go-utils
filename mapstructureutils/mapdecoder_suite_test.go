package mapdecoder_test

import (
	"testing"

	"go.uber.org/zap"

	"github.com/fgrosse/zaptest"
	"github.com/solo-io/go-utils/contextutils"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	. "github.com/onsi/gomega"
)

func TestMapDecoderServer(t *testing.T) {
	zaptest.Level = zap.InfoLevel
	logger := zaptest.LoggerWriter(GinkgoWriter)

	contextutils.SetFallbackLogger(logger.Sugar())

	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Map Decoder Suite", []Reporter{junitReporter})
}

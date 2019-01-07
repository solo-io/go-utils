package testutils

import (
	"github.com/fgrosse/zaptest"
	"github.com/solo-io/go-utils/pkg/contextutils"

	. "github.com/onsi/ginkgo"
)

func SetupLog() {
	logger := zaptest.LoggerWriter(GinkgoWriter)
	contextutils.SetFallbackLogger(logger.Sugar())
}

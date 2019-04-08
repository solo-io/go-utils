package logger_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	"github.com/solo-io/go-utils/logger"
	"go.uber.org/zap"
)

var _ = Describe("LogTest", func() {
	It("fallback logger works", func() {
		logger := logger.WithContext(context.TODO())
		logger.Infow("Testing", zap.Bool("isTest", true))
	})
})

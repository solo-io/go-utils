package loggingutils_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	"github.com/solo-io/go-utils/contextutils"
	. "github.com/solo-io/go-utils/loggingutils"
)

var _ = Describe("Grpc", func() {

	It("should inject log to context", func() {
		var logger zap.SugaredLogger
		var receivedLogger *zap.SugaredLogger

		f := LoggerInterceptor(&logger)
		f(context.Background(), "foo", nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			receivedLogger = contextutils.LoggerFrom(ctx)
			return nil, nil
		})
		Expect(receivedLogger).To(Equal(&logger))
	})

})

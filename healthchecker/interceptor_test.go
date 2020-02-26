package healthchecker_test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/healthchecker"
	"os"
	"syscall"
)


var _ = Describe("grpc healthchecker interceptor", func() {
	It("should inject log to context", func() {
		madeHealthCheckFail := make(chan struct{}, 1)
		sigs := make(chan os.Signal, 1)
		sigs<-syscall.SIGINT

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		f := healthchecker.GrpcUnaryServerHealthCheckerInterceptor(ctx, madeHealthCheckFail)
		f(context.Background(), "foo", nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			x, ok := <-madeHealthCheckFail
			Expect(ok).To(BeTrue())
			Expect(x).To(Equal(struct{}{}))
			return nil, nil
		})
	})

})

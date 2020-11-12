package healthchecker_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/healthchecker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ = Describe("grpc healthchecker interceptor", func() {
	It("should make the health check fail", func() {
		ctx, cancel := context.WithCancel(context.Background())
		stream := &mockServerTransportStream{}
		requestCtx := grpc.NewContextWithServerTransportStream(context.Background(), stream)
		cancel()
		f := healthchecker.GrpcUnaryServerHealthCheckerInterceptor(ctx)
		f(requestCtx, "foo", nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		expectedMd := metadata.MD{
			"x-envoy-immediate-health-check-fail": []string{""},
		}
		Expect(stream.header).To(Equal(expectedMd))
	})
	It("should not when context still alive", func() {
		stream := &mockServerTransportStream{}
		requestCtx := grpc.NewContextWithServerTransportStream(context.Background(), stream)
		f := healthchecker.GrpcUnaryServerHealthCheckerInterceptor(context.Background())
		f(requestCtx, "foo", nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		Expect(stream.header).To(BeNil())
	})

})

type mockServerTransportStream struct {
	header metadata.MD
}

func (m *mockServerTransportStream) Method() string {
	return ""
}

func (m *mockServerTransportStream) SetHeader(md metadata.MD) error {
	return nil
}

func (m *mockServerTransportStream) SendHeader(md metadata.MD) error {
	m.header = md
	return nil
}

func (m *mockServerTransportStream) SetTrailer(md metadata.MD) error {
	return nil
}

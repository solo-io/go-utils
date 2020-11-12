package healthchecker

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/solo-io/go-utils/contextutils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type grpcHealthChecker struct {
	srv  *health.Server
	name string
}

var _ HealthChecker = new(grpcHealthChecker)

func NewGrpc(serviceName string, grpcHealthServer *health.Server, failOnTerm bool) *grpcHealthChecker {
	hc := &grpcHealthChecker{}
	hc.name = serviceName

	hc.srv = grpcHealthServer
	hc.srv.SetServingStatus(serviceName, healthpb.HealthCheckResponse_SERVING)

	// TODO(yuval-k): we should remove this, as this shouldn't be done by this component, and
	// cannot be unit tested.
	// we can move this to a helper function if needed
	if failOnTerm {
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGTERM)

		go func() {
			<-sigterm
			hc.Fail()
		}()
	}

	return hc
}

func (hc *grpcHealthChecker) Fail() {
	hc.srv.SetServingStatus(hc.name, healthpb.HealthCheckResponse_NOT_SERVING)
}

func (hc *grpcHealthChecker) Ok() {
	hc.srv.SetServingStatus(hc.name, healthpb.HealthCheckResponse_SERVING)
}

func (hc *grpcHealthChecker) GetServer() *health.Server {
	return hc.srv
}

func GrpcUnaryServerHealthCheckerInterceptor(callerCtx context.Context) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		select {
		case <-callerCtx.Done():
			header := metadata.MD{"x-envoy-immediate-health-check-fail": []string{""}}
			err := grpc.SendHeader(ctx, header)
			logger := contextutils.LoggerFrom(ctx)
			logger.Debugf("received signal that caller context has been canceled. Sending header %v %v", header, err)
		default:
		}
		return handler(ctx, req)
	}
}

package healthchecker

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type grpcHealthChecker struct {
	srv  *health.Server
	ok   uint32
	name string
}

var _ HealthChecker = new(grpcHealthChecker)

func NewGrpc(serviceName string, grpcHealthServer *health.Server) *grpcHealthChecker {
	hc := &grpcHealthChecker{}
	hc.ok = 1
	hc.name = serviceName

	hc.srv = grpcHealthServer
	hc.srv.SetServingStatus(serviceName, healthpb.HealthCheckResponse_SERVING)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)

	go func() {
		<-sigterm
		atomic.StoreUint32(&hc.ok, 0)
		hc.srv.SetServingStatus(serviceName, healthpb.HealthCheckResponse_NOT_SERVING)
	}()

	return hc
}

func (hc *grpcHealthChecker) Fail() {
	atomic.StoreUint32(&hc.ok, 0)
	hc.srv.SetServingStatus(hc.name, healthpb.HealthCheckResponse_NOT_SERVING)
}

func (hc *grpcHealthChecker) GetServer() *health.Server {
	return hc.srv
}

func GrpcUnaryServerHealthCheckerInterceptor(callerCtx context.Context, failedHealthCheck chan struct{}) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger := contextutils.LoggerFrom(ctx)
		logger.Debugw("Intercepted request: ", zap.Any("req", req))
		logger.Debugw("Intercepted request info: ", zap.Any("info", info))

		select {
		case <-callerCtx.Done():
			header := metadata.Pairs("x-envoy-immediate-health-check-fail", "")
			grpc.SendHeader(ctx, header)
			logger.Debugf("received signal that caller context has been canceled. Sending header %v", header)
			failedHealthCheck <- struct{}{}
		default:
		}

		resp, err := handler(ctx, req)
		logger.Debugw("Intercepted response: ", zap.Any("resp", resp))
		logger.Debugw("Intercepted response error: ", zap.Error(err))
		return resp, err
	}
}
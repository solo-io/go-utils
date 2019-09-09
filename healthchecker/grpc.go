package healthchecker

import (
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

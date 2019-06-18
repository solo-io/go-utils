package envutils

import (
	"context"
	"os"
	"strconv"

	"go.uber.org/zap"

	"github.com/solo-io/go-utils/contextutils"
)

func MustGetPodNamespace(ctx context.Context) string {
	contextutils.LoggerFrom(ctx).Infow("Looking for install namespace in POD_NAMESPACE environment variable")
	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		contextutils.LoggerFrom(ctx).Fatalw("Could not determine namespace, must have non-empty POD_NAMESPACE in environment")
	}
	contextutils.LoggerFrom(ctx).Infow("Found install namespace", zap.String("installNamespace", namespace))
	return namespace
}

func MustGetGrpcPort(ctx context.Context) int {
	contextutils.LoggerFrom(ctx).Infow("Looking for grpc port in GRPC_PORT environment variable")
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		contextutils.LoggerFrom(ctx).Fatalw("Could not determine port, must have non-empty GRPC_PORT in environment")
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Failed to parse grpc port",
			zap.Error(err),
			zap.String("port", port))
	}
	contextutils.LoggerFrom(ctx).Infow("Found grpc port", zap.Int("grpcPort", p))
	return p
}

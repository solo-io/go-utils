package loggingutils

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func LoggerInterceptor(logger *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = contextutils.WithExistingLogger(ctx, logger)
		return handler(ctx, req)
	}
}

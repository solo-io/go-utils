package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerKey struct{}

// This logger is used when there is no logger attached to the context.
// Rather than returning nil and causing a panic, we will use the fallback
// logger. Fallback logger is tagged with logger=fallback to make sure
// that code that doesn't set the logger correctly can be caught at runtime.
var fallbackLogger *zap.SugaredLogger

func init() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if logger, err := config.Build(); err != nil {

		// We failed to create a fallback logger. Our fallback
		// unfortunately falls back to noop.
		fallbackLogger = zap.NewNop().Sugar()
	} else {
		fallbackLogger = logger.Sugar()
	}
}

func WithLogger(ctx context.Context, log *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey{}, log)
}

func NewContext(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, loggerKey{}, WithContext(ctx).With(fields))
}

func WithContext(ctx context.Context) *zap.SugaredLogger {
	if ctx == nil {
		return fallbackLogger
	}
	if ctxLogger, ok := ctx.Value(loggerKey{}).(*zap.SugaredLogger); ok {
		return ctxLogger
	} else {
		return fallbackLogger
	}
}

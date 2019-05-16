/*
copied from: github.com/knative/serving/pkg/logging/logger.go
Copyright 2018 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package contextutils

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerKey struct{}

var (
	// This logger is used when there is no logger attached to the context.
	// Rather than returning nil and causing a panic, we will use the fallback
	// logger.
	fallbackLogger *zap.SugaredLogger
	// The atomic level set for any logger built here. Accessing this atomic level
	// and calling set level will change the log output dynamically.
	level zap.AtomicLevel
)

func buildLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	level = zap.NewAtomicLevel()
	config.Level = level
	return config.Build()
}

func init() {
	if logger, err := buildLogger(); err != nil {

		// We failed to create a fallback logger. Our fallback
		// unfortunately falls back to noop.
		fallbackLogger = zap.NewNop().Sugar()
	} else {
		fallbackLogger = logger.Sugar()
	}
}

func SetFallbackLogger(logger *zap.SugaredLogger) {
	fallbackLogger = logger
}

// WithLogger returns a copy of parent context in which the
// value associated with logger key is the supplied logger.
func withLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// FromContext returns the logger stored in context.
// Returns nil if no logger is set in context, or if the stored value is
// not of correct type.
func fromContext(ctx context.Context) *zap.SugaredLogger {
	if ctx != nil {
		if logger, ok := ctx.Value(loggerKey{}).(*zap.SugaredLogger); ok {
			return logger
		}
	}
	return fallbackLogger
}

func SetLogLevel(l zapcore.Level) {
	level.SetLevel(l)
}

func GetLogLevel() zapcore.Level {
	return level.Level()
}

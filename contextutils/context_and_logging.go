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
	"os"

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

const LogLevelEnvName = "LOG_LEVEL"

func buildProductionLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	level = zap.NewAtomicLevel()
	config.Level = level
	return config.Build()
}

func buildSplitOutputProductionLogger() (*zap.Logger, error) {
	logger, err := buildProductionLogger()
	if err != nil {
		return nil, err
	}

	// Define level-handling logic.
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	// High-priority output should go to standard error, and low-priority
	// output should go to standard out.
	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	buildEncoder := func() zapcore.Encoder {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		return zapcore.NewJSONEncoder(encoderConfig)
	}

	splitOutput := zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		// Join the outputs, encoders, and level-handling functions into
		// zapcore.Cores, then tee the cores together.
		return zapcore.NewTee(
			zapcore.NewCore(
				buildEncoder(), consoleErrors, highPriority,
			),
			zapcore.NewCore(
				buildEncoder(), consoleDebugging, lowPriority,
			),
		)
	})

	return logger.WithOptions(splitOutput, zap.AddCaller(), zap.AddStacktrace(highPriority)), nil
}

func init() {
	buildLogger := buildProductionLogger
	// Specify gloo.splitLogOutput=true when installing Gloo via Helm to set the SPLIT_LOG_OUTPUT env var
	if os.Getenv("SPLIT_LOG_OUTPUT") == "true" {
		buildLogger = buildSplitOutputProductionLogger
	}
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

func SetLogLevelFromString(logLevel string) {
	var setLevel zapcore.Level

	switch logLevel {
	case "debug":
		setLevel = zapcore.DebugLevel
	case "warn":
		setLevel = zapcore.WarnLevel
	case "error":
		setLevel = zapcore.ErrorLevel
	case "panic":
		setLevel = zapcore.PanicLevel
	case "fatal":
		setLevel = zapcore.FatalLevel
	default:
		setLevel = zapcore.InfoLevel
	}

	SetLogLevel(setLevel)
}

func SetLogLevel(l zapcore.Level) {
	level.SetLevel(l)
}

func GetLogHandler() zap.AtomicLevel {
	return level
}

func GetLogLevel() zapcore.Level {
	return level.Level()
}

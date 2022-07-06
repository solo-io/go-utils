package stats_test

import (
	"context"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/stats"
	"github.com/solo-io/go-utils/testutils/goimpl"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ = Describe("Stats", func() {

	Context("StartStatsSeverWithPort", func() {

		var (
			ctx    context.Context
			cancel context.CancelFunc

			resetEnvVars func()
		)

		BeforeEach(func() {
			ctx, cancel = context.WithCancel(context.Background())
		})

		AfterEach(func() {
			cancel()

			resetEnvVars()
		})

		When("StartOptions are default", func() {

			var (
				startupOptions stats.StartupOptions
			)

			BeforeEach(func() {
				startupOptions = stats.DefaultStartupOptions()

				originalEnvValue := os.Getenv(startupOptions.EnvVar)
				err := os.Setenv(startupOptions.EnvVar, startupOptions.EnabledValue)
				Expect(err).NotTo(HaveOccurred())

				resetEnvVars = func() {
					err := os.Setenv(startupOptions.EnvVar, originalEnvValue)
					Expect(err).NotTo(HaveOccurred())
				}

				stats.StartCancellableStatsServerWithPort(ctx, startupOptions)
			})

			It("can handle requests to /logging", func() {
				response, err := goimpl.Curl(fmt.Sprintf("http://localhost:%d/logging", startupOptions.Port))
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(Equal("{\"level\":\"info\"}\n"))
			})
		})

		When("StartOptions are default and LOG_LEVEL set", func() {

			var (
				startupOptions stats.StartupOptions
			)

			BeforeEach(func() {
				startupOptions = stats.DefaultStartupOptions()

				originalEnvValue := os.Getenv(startupOptions.EnvVar)
				err := os.Setenv(startupOptions.EnvVar, startupOptions.EnabledValue)
				Expect(err).NotTo(HaveOccurred())

				originalLogLevel := os.Getenv(contextutils.LogLevelEnvName)
				err = os.Setenv(contextutils.LogLevelEnvName, "error")
				Expect(err).NotTo(HaveOccurred())

				resetEnvVars = func() {
					err = os.Setenv(startupOptions.EnvVar, originalEnvValue)
					Expect(err).NotTo(HaveOccurred())

					err = os.Setenv(contextutils.LogLevelEnvName, originalLogLevel)
					Expect(err).NotTo(HaveOccurred())
				}

				stats.StartCancellableStatsServerWithPort(ctx, startupOptions)
			})

			It("can handle requests to /logging", func() {
				response, err := goimpl.Curl(fmt.Sprintf("http://localhost:%d/logging", startupOptions.Port))
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(Equal("{\"level\":\"error\"}\n"))
			})
		})

		When("StartOptions.LogLevel is set", func() {

			var (
				startupOptions stats.StartupOptions
			)

			BeforeEach(func() {
				startupOptions = stats.DefaultStartupOptions()
				customLogLevel := zap.NewAtomicLevelAt(zapcore.DebugLevel)
				startupOptions.LogLevel = &customLogLevel

				originalEnvValue := os.Getenv(startupOptions.EnvVar)
				err := os.Setenv(startupOptions.EnvVar, startupOptions.EnabledValue)
				Expect(err).NotTo(HaveOccurred())

				resetEnvVars = func() {
					err := os.Setenv(startupOptions.EnvVar, originalEnvValue)
					Expect(err).NotTo(HaveOccurred())
				}

				stats.StartCancellableStatsServerWithPort(ctx, startupOptions)
			})

			It("can handle requests to /logging", func() {
				response, err := goimpl.Curl(fmt.Sprintf("http://localhost:%d/logging", startupOptions.Port))
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(Equal("{\"level\":\"debug\"}\n"))
			})

		})

	})

})

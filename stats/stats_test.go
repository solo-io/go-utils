package stats_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
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
				getLogLevelRequest, err := buildGetLogLevelRequest(startupOptions.Port)
				Expect(err).NotTo(HaveOccurred())

				response, err := goimpl.ExecuteRequest(getLogLevelRequest)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(Equal("{\"level\":\"info\"}\n"))
			})

			It("can use /logging to change the level, without restarting", func() {
				// initially the logLevel = INFO
				getLogLevelRequest, err := buildGetLogLevelRequest(startupOptions.Port)
				Expect(err).NotTo(HaveOccurred())

				response, err := goimpl.ExecuteRequest(getLogLevelRequest)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(Equal("{\"level\":\"info\"}\n"))

				// then it's updated to DEBUG
				setLogLevelRequest, err := buildSetLogLevelRequest(startupOptions.Port, "debug")
				Expect(err).NotTo(HaveOccurred())

				response, err = goimpl.ExecuteRequest(setLogLevelRequest)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(Equal("{\"level\":\"debug\"}\n"))

				// we confirm that we return DEBUG moving forward
				getLogLevelRequest, err = buildGetLogLevelRequest(startupOptions.Port)
				Expect(err).NotTo(HaveOccurred())

				response, err = goimpl.ExecuteRequest(getLogLevelRequest)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(Equal("{\"level\":\"debug\"}\n"))
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
				getLogLevelRequest, err := buildGetLogLevelRequest(startupOptions.Port)
				Expect(err).NotTo(HaveOccurred())

				response, err := goimpl.ExecuteRequest(getLogLevelRequest)
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
				getLogLevelRequest, err := buildGetLogLevelRequest(startupOptions.Port)
				Expect(err).NotTo(HaveOccurred())

				response, err := goimpl.ExecuteRequest(getLogLevelRequest)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(Equal("{\"level\":\"debug\"}\n"))
			})

		})

	})

})

func buildGetLogLevelRequest(port int) (*http.Request, error) {
	url := fmt.Sprintf("http://localhost:%d/logging", port)
	body := bytes.NewReader([]byte(url))

	return http.NewRequest("GET", url, body)
}

func buildSetLogLevelRequest(port int, newLevel string) (*http.Request, error) {
	url := fmt.Sprintf("http://localhost:%d/logging", port)
	body := bytes.NewReader([]byte(fmt.Sprintf("{\"level\": \"%s\"}", newLevel)))

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Context-Type", "application/json")

	return req, nil
}

package stats_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/onsi/ginkgo/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/stats"
	"github.com/solo-io/go-utils/testutils/goimpl"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var statsServerPort = int32(stats.DefaultPort)

func NextStatsBindPort() int {
	return int(AdvanceBindPort(&statsServerPort))
}

func AdvanceBindPort(p *int32) int32 {
	return atomic.AddInt32(p, 1) + int32(config.GinkgoConfig.ParallelNode)
}

var _ = Describe("Stats", func() {

	Context("StartStatsSeverWithPort", func() {

		var (
			ctx    context.Context
			cancel context.CancelFunc
		)

		BeforeEach(func() {
			ctx, cancel = context.WithCancel(context.Background())

			// Tests in this suite expect the log level to be INFO to start
			contextutils.SetLogLevel(zapcore.InfoLevel)

			err := os.Unsetenv(contextutils.LogLevelEnvName)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			cancel()
		})

		When("StartOptions are default", func() {

			var (
				startupOptions stats.StartupOptions
			)

			BeforeEach(func() {
				startupOptions = stats.DefaultStartupOptions()
				startupOptions.Port = NextStatsBindPort()

				stats.StartCancellableStatsServerWithPort(ctx, startupOptions)
			})

			It("can handle requests to /logging", func() {
				By("GET request to /logging to get log level")
				getLogLevelRequest, err := buildGetLogLevelRequest(startupOptions.Port)
				Expect(err).NotTo(HaveOccurred())

				response, err := goimpl.ExecuteRequest(getLogLevelRequest)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(Equal("{\"level\":\"info\"}\n"))

				By("PUT request to /logging to change log level ")
				setLogLevelRequest, err := buildSetLogLevelRequest(startupOptions.Port, zapcore.DebugLevel)
				Expect(err).NotTo(HaveOccurred())

				response, err = goimpl.ExecuteRequest(setLogLevelRequest)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(Equal("{\"level\":\"debug\"}\n"))

				By("GET request to /logging to confirm it returns new log level")
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
				startupOptions.Port = NextStatsBindPort()

				err := os.Setenv(contextutils.LogLevelEnvName, zapcore.ErrorLevel.String())
				Expect(err).NotTo(HaveOccurred())

				stats.StartCancellableStatsServerWithPort(ctx, startupOptions)
			})

			It("can handle requests to /logging", func() {
				By("GET request to /logging to confirm it returns value set by LOG_LEVEL")
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
				startupOptions.Port = NextStatsBindPort()
				customLogLevel := zap.NewAtomicLevelAt(zapcore.DebugLevel)
				startupOptions.LogLevel = &customLogLevel

				stats.StartCancellableStatsServerWithPort(ctx, startupOptions)
			})

			It("can handle requests to /logging", func() {
				By("GET request to /logging to confirm it returns value set by StartupOptions.LogLevel")
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

func buildSetLogLevelRequest(port int, level zapcore.Level) (*http.Request, error) {
	url := fmt.Sprintf("http://localhost:%d/logging", port)
	body := bytes.NewReader([]byte(fmt.Sprintf("{\"level\": \"%s\"}", level.String())))

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Context-Type", "application/json")

	return req, nil
}

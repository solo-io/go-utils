package stats_test

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
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

			startupOptions stats.StartupOptions
		)

		BeforeEach(func() {
			ctx, cancel = context.WithCancel(context.Background())

			// Tests in this suite expect the log level to be INFO to start
			contextutils.SetLogLevel(zapcore.InfoLevel)

			err := os.Unsetenv(contextutils.LogLevelEnvName)
			Expect(err).NotTo(HaveOccurred())

			// Initialize startupOptions to default value
			startupOptions = stats.DefaultStartupOptions()
		})

		AfterEach(func() {
			cancel()

			// Ensure that after we cancel the context, which initiates a shutdown of the server,
			// that we wait for the port to be released, so we can start up a next server on the subsequent test
			EventuallyPortAvailable(startupOptions.Port)
		})

		When("StartOptions are default", func() {

			BeforeEach(func() {
				startupOptions = stats.DefaultStartupOptions()

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

			BeforeEach(func() {
				startupOptions = stats.DefaultStartupOptions()

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

			BeforeEach(func() {
				startupOptions = stats.DefaultStartupOptions()
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

func EventuallyPortAvailable(port int) {
	EventuallyWithOffset(1, func() error {
		timeout := time.Millisecond * 100
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", fmt.Sprintf("%d", port)), timeout)
		if err != nil {
			// nothing listening on this port, we can proceed because the server has been shutdown
			return nil
		}
		_ = conn.Close()
		return eris.New(fmt.Sprintf("connection still open on port %d, expected it to be closed", port))
	}, time.Second*3, time.Millisecond*100).ShouldNot(HaveOccurred())
}

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

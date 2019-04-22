package clicore

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Glooshot CLI", func() {

	var standardCobraHelpBlockMatcher = MatchRegexp("Available Commands:")

	BeforeEach(func() {
	})

	Context("basic args and flags", func() {
		It("should return help messages without error", func() {
			_, _, err := appWithSimpleOutput("-h")
			Expect(err).NotTo(HaveOccurred())
			_, _, err = appWithSimpleOutput("help")
			Expect(err).NotTo(HaveOccurred())
			_, _, err = appWithSimpleOutput("--help")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("expect human-friendly errors", func() {
		It("should return human-friendly errors on bad input", func() {
			cliOut := appWithLoggerOutput("--h")
			Expect(cliOut.CobraStdout).To(Equal(""))
			Expect(cliOut.CobraStderr).To(standardCobraHelpBlockMatcher)
			// logs are not used in this code path so they should be empty
			Expect(cliOut.LoggerConsoleStout).To(Equal(""))
			Expect(cliOut.LoggerConsoleStderr).To(Equal(""))
		})
	})

	Context("expect human-friendly logs", func() {
		It("should return human-friendly errors on bad input", func() {
			cliOut := appWithLoggerOutput("--temp")
			Expect(cliOut.CobraStdout).
				To(Equal("cobra says 'hisssss' - but he should leave the console logs to the CliLog* utils."))
			Expect(cliOut.CobraStderr).
				To(MatchRegexp("Error: cobra says 'hisssss' again - it's ok because this is a passed error"))
			Expect(cliOut.CobraStderr).
				To(standardCobraHelpBlockMatcher)
			Expect(cliOut.LoggerConsoleStout).
				To(Equal(`this info log should go to file and console
this warn log should go to file and console
this infow log should go to file and console
this warnw log should go to file and console
`))
			Expect(cliOut.LoggerConsoleStderr).To(Equal(`this error log should go to file and console
this errorw log should go to file and console
`))
			// match the tags that are part of the rich log output
			Expect(cliOut.LoggerFileContent).To(MatchRegexp("level"))
			Expect(cliOut.LoggerFileContent).To(MatchRegexp("ts"))
			Expect(cliOut.LoggerFileContent).To(MatchRegexp("warn"))
			Expect(cliOut.LoggerFileContent).To(MatchRegexp("error"))
			Expect(cliOut.LoggerFileContent).To(MatchRegexp("dev"))
			Expect(cliOut.LoggerFileContent).To(MatchRegexp("msg"))
			Expect(cliOut.LoggerFileContent).To(MatchRegexp("logger"))
			// match (or not) the fragments that we get in the console. Using regex since timestamp is random
			// see sampleLogFileContent for an example of the full output
			Expect(cliOut.LoggerFileContent).NotTo(MatchRegexp("CliLog* utils"))
			Expect(cliOut.LoggerFileContent).To(MatchRegexp("ok because this is a passed error"))
			Expect(cliOut.LoggerFileContent).To(MatchRegexp("info log"))
			Expect(cliOut.LoggerFileContent).To(MatchRegexp("warn log"))
			Expect(cliOut.LoggerFileContent).To(MatchRegexp("error log"))
			for i := 1; i <= 3; i++ {
				Expect(cliOut.LoggerFileContent).To(MatchRegexp(fmt.Sprintf("extrakey%v", i)))
				Expect(cliOut.LoggerFileContent).To(MatchRegexp(fmt.Sprintf("val%v", i)))
			}

		})

	})
})

func appWithSimpleOutput(args string) (string, string, error) {
	co := appWithLoggerOutput(args)
	return co.CobraStdout, co.CobraStderr, nil
}

// This is all you need to do to use the cli logger in a test environment
func appWithLoggerOutput(args string) CliOutput {
	cliOutput, err := sampleAppConfig.RunForTest(args)
	Expect(err).NotTo(HaveOccurred())
	return cliOutput
}

var (
	appVersion           = "test"
	fileLogPathElements  = []string{".sample", "log", "dir"}
	outputModeEnvVar     = "SET_OUTPUT_MODE"
	errorMessagePreamble = "error running cli"
)

func SampleCobraCli(ctx context.Context, version string) *cobra.Command {
	cmd := &cobra.Command{}
	return cmd
}

var sampleAppConfig = CommandConfig{
	Command:             SampleCobraCli,
	Version:             appVersion,
	FileLogPathElements: fileLogPathElements,
	OutputModeEnvVar:    outputModeEnvVar,
	RootErrorMessage:    errorMessagePreamble,
	LoggingContext:      []interface{}{"version", appVersion},
}

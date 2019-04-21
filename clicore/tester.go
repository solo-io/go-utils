package clicore

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type CliTestConfig struct {
	CommandConfig *CommandConfig
	MockTargets   *MockTargets
	TestArgs      string
	preparedCmd   *cobra.Command
	ctx           context.Context
}

// CliOutput captures all the relevant output from a Cobra Command
// For clarity and simplicity, output from zapcore loggers are stored separately
// otherwise, it would be necessary to coordinate the initialization of the loggers
// with the os.Std*** manipulation done in ExecuteCliOutErr
type CliOutput struct {
	LoggerConsoleStderr string
	LoggerConsoleStout  string
	LoggerFileContent   string
	CobraStderr         string
	CobraStdout         string
}

// ExecuteCliOutErr is a helper for calling a cobra command within a test
// handleCommandError is an optional parameter that can be used for replicating the error handler that you use
// when calling the command from your main file. This overcomes the chicken-and-egg problem of calling os.Exit on
// CLI errors. Suggestion: duplicate the error handling used when calling command.Execute(), but replace fatal logging
// with a non-fatal equivalent
func ExecuteCliOutErr(ct *CliTestConfig) (string, string, error) {
	stdOut := os.Stdout
	stdErr := os.Stderr
	r1, w1, err := os.Pipe()
	if err != nil {
		return "", "", err
	}
	r2, w2, err := os.Pipe()
	if err != nil {
		return "", "", err
	}
	os.Stdout = w1
	os.Stderr = w2

	ct.preparedCmd.SetArgs(strings.Split(ct.TestArgs, " "))
	commandErr := ct.preparedCmd.Execute()
	if commandErr != nil {
		// This error handler has been specified to match the Fatalw handler used in the binary.
		// With the important difference that it does not call os.Exit.
		contextutils.LoggerFrom(ct.ctx).Errorw(ct.CommandConfig.RootErrorMessage, zap.Error(commandErr))
	}

	chan1 := make(chan string)
	chan2 := make(chan string)

	chan1err := make(chan error)
	chan2err := make(chan error)

	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r1)
		chan1err <- err
		chan1 <- buf.String()
	}()
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r2)
		chan2err <- err
		chan2 <- buf.String()
	}()

	// back to normal state
	os.Stdout = stdOut // restoring the real stdout
	os.Stderr = stdErr
	if err := w1.Close(); err != nil {
		return "", "", err
	}
	if err := w2.Close(); err != nil {
		return "", "", err
	}
	if err := <-chan1err; err != nil {
		return "", "", err
	}
	if err := <-chan2err; err != nil {
		return "", "", err
	}
	capturedStdout := <-chan1
	capturedStderr := <-chan2

	return strings.TrimSuffix(capturedStdout, "\n"),
		strings.TrimSuffix(capturedStderr, "\n"),
		err
}

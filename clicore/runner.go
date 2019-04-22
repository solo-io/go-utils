package clicore

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type RootCommandFunc func(ctx context.Context, version string) (rootCmd *cobra.Command)
type CommandConfig struct {
	Args                string
	Command             RootCommandFunc
	CommandErrorHandler func(error)
	RootErrorMessage    string
	OutputModeEnvVar    string
	LoggingContext      []interface{}
	FileLogPathElements []string
	Version             string
	ctx                 context.Context
}

func (cc CommandConfig) Run() {
	cmd := cc.setInitialContextAndSetLoggerFromConfig(nil).prepareCommand()
	if err := cmd.Execute(); err != nil {
		contextutils.LoggerFrom(cc.ctx).Fatalw(cc.RootErrorMessage, zap.Error(err))
	}
}

func (cc *CommandConfig) RunForTest(args string) (CliOutput, error) {
	return cc.setContextAndPrepareCommandForTest(args).callCobraCommandForTest()
}

func (cc *CommandConfig) setContextAndPrepareCommandForTest(args string) *CliTestConfig {
	mockTargets := NewMockTargets()
	cc.setInitialContextAndSetLoggerFromConfig(&mockTargets)
	return &CliTestConfig{
		CommandConfig: cc,
		MockTargets:   &mockTargets,
		TestArgs:      args,
		preparedCmd:   cc.prepareCommand(),
		ctx:           cc.ctx,
	}
}

func (ct *CliTestConfig) callCobraCommandForTest() (CliOutput, error) {
	cliOut := CliOutput{}
	var err error
	cliOut.CobraStdout, cliOut.CobraStderr, err = ExecuteCliOutErr(ct)
	// After the command has been executed, there should be content in the logs
	cliOut.LoggerConsoleStout, _, _ = ct.MockTargets.Stdout.Summarize()
	cliOut.LoggerConsoleStderr, _, _ = ct.MockTargets.Stderr.Summarize()
	cliOut.LoggerFileContent, _, _ = ct.MockTargets.FileLog.Summarize()
	return cliOut, err
}

func (cc *CommandConfig) setInitialContextAndSetLoggerFromConfig(mockTargets *MockTargets) *CommandConfig {
	cliLogger := &zap.SugaredLogger{}
	if mockTargets == nil {
		cliLogger = BuildCliLogger(cc.FileLogPathElements, cc.OutputModeEnvVar)
	} else {
		cliLogger = BuildMockedCliLogger(cc.FileLogPathElements, cc.OutputModeEnvVar, mockTargets)
	}
	contextutils.SetFallbackLogger(cliLogger)
	ctx := contextutils.WithLogger(context.Background(), cc.Version)
	cc.ctx = contextutils.WithLoggerValues(ctx, cc.LoggingContext...)
	return cc
}

func (cc CommandConfig) prepareCommand() *cobra.Command {
	return cc.Command(cc.ctx, cc.Version)
}

package commands

import (
	"bytes"
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/solo-io/go-utils/securityscanutils/internal"
	"github.com/spf13/cobra"
)

// Configure the CLI, including possible commands and input args.
func RootCommand(ctx context.Context) *cobra.Command {
	globalFlags := &internal.GlobalFlags{}

	cmd := &cobra.Command{
		Use:   "cvectl [command]",
		Short: "CLI for identifying CVEs in images",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if globalFlags.Verbose {
				logrus.SetLevel(logrus.DebugLevel)
			}
		},
		SilenceErrors: true,
	}

	// Use custom logrus formatter
	logrus.SetFormatter(logFormatter{})

	// set global CLI flags
	globalFlags.AddToFlags(cmd.PersistentFlags())

	cmd.AddCommand(
		ScanRepoCommand(ctx, globalFlags),
		ScanVersionCommand(ctx, globalFlags),

		FormatResultsCommand(ctx, globalFlags))

	return cmd
}

type logFormatter struct{}

func (logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var buf bytes.Buffer
	switch entry.Level {
	case logrus.DebugLevel:
		_, _ = color.New(color.FgRed).Fprintln(&buf, entry.Message)
	case logrus.InfoLevel:
		_, _ = fmt.Fprintln(&buf, entry.Message)
	case logrus.WarnLevel:
		_, _ = color.New(color.FgYellow).Fprint(&buf, "warning: ")
		_, _ = fmt.Fprintln(&buf, entry.Message)
	case logrus.ErrorLevel:
		_, _ = color.New(color.FgRed).Fprint(&buf, "error: ")
		_, _ = fmt.Fprintln(&buf, entry.Message)
	case logrus.FatalLevel:
		_, _ = color.New(color.FgRed).Fprint(&buf, "fatal: ")
		_, _ = fmt.Fprintln(&buf, entry.Message)
	case logrus.PanicLevel:
		_, _ = color.New(color.FgRed).Fprint(&buf, "panic: ")
		_, _ = fmt.Fprintln(&buf, entry.Message)
	}

	return buf.Bytes(), nil
}

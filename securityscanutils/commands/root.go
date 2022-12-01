package commands

import (
	"bytes"
	"context"
	"fmt"

	"github.com/spf13/pflag"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// RootCommand configure the CLIs, including possible commands and input args.
func RootCommand(ctx context.Context) *cobra.Command {
	rootOptions := &RootOptions{}

	cmd := &cobra.Command{
		Use:   "cvectl [command]",
		Short: "CLI for identifying CVEs in images",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if rootOptions.Verbose {
				logrus.SetLevel(logrus.DebugLevel)
			}
		},
		SilenceErrors: true,
	}

	// Use custom logrus formatter
	logrus.SetFormatter(logFormatter{})

	// set global CLI flags
	rootOptions.AddToFlags(cmd.PersistentFlags())

	cmd.AddCommand(
		ScanRepoCommand(ctx, rootOptions),
		ScanVersionCommand(ctx, rootOptions),

		FormatResultsCommand(ctx, rootOptions))

	return cmd
}

type RootOptions struct {
	Verbose bool
}

func (r *RootOptions) AddToFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&r.Verbose, "verbose", "v", false, "Enable verbose logging")
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

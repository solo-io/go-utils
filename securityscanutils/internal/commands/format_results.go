package commands

import (
    "context"
    "github.com/solo-io/go-utils/securityscanutils/internal"
    "github.com/spf13/cobra"
    "github.com/spf13/pflag"
)

func FormatResultsCommand(ctx context.Context, globalFlags *internal.GlobalFlags) *cobra.Command {
    opts := &formatResultsOptions{
        GlobalFlags: globalFlags,
    }

    cmd := &cobra.Command{
        Use:   "format-results",
        Aliases: []string{"format", "f"},
        Short: "[TODO]",
        RunE: func(cmd *cobra.Command, args []string) error {
            return doFormatResults(ctx, opts)
        },
    }
    opts.addToFlags(cmd.Flags())
    cmd.SilenceUsage = true
    return cmd
}

type formatResultsOptions struct {
    *internal.GlobalFlags
}

func (f *formatResultsOptions) addToFlags(flags *pflag.FlagSet) {
}

func doFormatResults(ctx context.Context, opts *formatResultsOptions) error {
    return nil
}
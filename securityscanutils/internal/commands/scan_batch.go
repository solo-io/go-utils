package commands

import (
    "context"
    "github.com/solo-io/go-utils/securityscanutils/internal"
    "github.com/spf13/cobra"
    "github.com/spf13/pflag"
)

func ScanBatchCommand(ctx context.Context, globalFlags *internal.GlobalFlags) *cobra.Command {
    opts := &scanBatchOptions{
        GlobalFlags: globalFlags,
    }

    cmd := &cobra.Command{
        Use:   "scan-batch",
        Aliases: []string{"batch"},
        Short: "Execute a bulk scan of multiple images/versions",
        RunE: func(cmd *cobra.Command, args []string) error {
            return doScanBatch(ctx, opts)
        },
    }
    opts.addToFlags(cmd.Flags())
    cmd.SilenceUsage = true
    return cmd
}

func doScanBatch(ctx context.Context, opts *scanBatchOptions) error {
    return nil
}

type scanBatchOptions struct {
    *internal.GlobalFlags
}

func (m *scanBatchOptions) addToFlags(flags *pflag.FlagSet) {
}
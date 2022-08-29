package securityscanutils

import (
    "context"
    "github.com/solo-io/go-utils/securityscanutils/internal/commands"
    "github.com/spf13/cobra"
)

func RootCommand(ctx context.Context) *cobra.Command {
    return commands.RootCommand(ctx)
}
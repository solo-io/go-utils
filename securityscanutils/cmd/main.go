package main

import (
    "context"
    "github.com/solo-io/go-utils/securityscanutils/internal/commands"
    "log"
)

func main() {
    ctx := context.Background()
    if err := commands.RootCommand(ctx).Execute(); err != nil {
        log.Fatalf("unable to run: %v\n", err)
    }
}

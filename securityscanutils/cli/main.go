package main

import (
    "context"
    "github.com/solo-io/go-utils/securityscanutils/commands"
    "log"
)

func main() {
    ctx := context.Background()

    cmd := commands.RootCommand(ctx)
    if err := cmd.Execute(); err != nil {
        log.Fatalf("unable to run: %v\n", err)
    }
}

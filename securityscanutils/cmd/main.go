package main

import (
"context"
    "github.com/solo-io/go-utils/securityscanutils"
    "log"
)

func main() {
    ctx := context.Background()
    if err := securityscanutils.RootCommand(ctx).Execute(); err != nil {
        log.Fatalf("unable to run: %v\n", err)
    }
}

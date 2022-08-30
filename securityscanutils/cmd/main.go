package main

import (
	"context"
	"log"

	"github.com/solo-io/go-utils/securityscanutils/internal/commands"
)

func main() {
	ctx := context.Background()
	if err := commands.RootCommand(ctx).Execute(); err != nil {
		log.Fatalf("unable to run: %v\n", err)
	}
}

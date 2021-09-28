package main

import (
	"context"
	"log"

	"github.com/solo-io/go-utils/securityscanutils"
)

func main() {
	ctx := context.Background()
	app := securityscanutils.RootApp(ctx)
	if err := app.Execute(); err != nil {
		log.Fatalf("unable to run: %v\n", err)
	}
}

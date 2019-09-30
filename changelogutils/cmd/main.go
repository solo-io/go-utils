package main

import (
	"context"
	"os"

	"github.com/solo-io/go-utils/changelogutils"
	"github.com/solo-io/go-utils/log"
)

// This is a reference script, showing how to call the generation script.
// See the README.md file for an example of the output
func main() {
	ctx := context.Background()
	repoRootPath := "."
	owner := "solo-io"
	repo := "go-utils"
	changelogDirPath := "changelog"

	// consider writing to stdout to enhance makefile/io readability `go run cmd/main.go > changelogSummary.md`
	w := os.Stdout
	err := changelogutils.GenerateChangelogFromLocalDirectory(ctx, repoRootPath, owner, repo, changelogDirPath, w)
	if err != nil {
		log.Fatalf("unable to run: %v", err)
	}
}

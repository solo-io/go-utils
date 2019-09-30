package main

import (
	"context"
	"os"

	"github.com/solo-io/go-utils/changelogutils"
	"github.com/solo-io/go-utils/log"
)

func main() {
	ctx := context.Background()
	repoRootPath := "."
	owner := "solo-io"
	repo := "go-utils"
	sha := "n/a"
	changelogDirPath := "changelog"

	w := os.Stdout
	err := changelogutils.GenerateChangelogFromLocalDirectory(ctx, repoRootPath, owner, repo, sha, changelogDirPath, w)
	if err != nil {
		log.Fatalf("unable to run: %v", err)
	}
}

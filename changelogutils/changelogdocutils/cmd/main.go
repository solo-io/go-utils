package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v32/github"
	"github.com/solo-io/go-utils/changelogutils/changelogdocutils"
	"golang.org/x/oauth2"
	"os"
)

func main() {
	ctx := context.Background()
	if os.Getenv("GITHUB_TOKEN") == "" {
		fmt.Println("SET GITHUB_TOKEN")
		os.Exit(1)
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	gen := changelogdocutils.NewMergedReleaseGenerator(client, "solo-io", "solo-projects", "gloo")
	changelog, err := gen.Generate(ctx)
	if err != nil {
		fmt.Println("error", err.Error())
		os.Exit(1)
	}
	f, err := os.Create("./tmp2.md")
	if err != nil {
		fmt.Println(err.Error())
	}
	f.WriteString(changelog)
	//fmt.Println(changelog)
}
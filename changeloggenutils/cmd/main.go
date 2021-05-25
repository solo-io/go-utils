package main

import (
	"context"
	"fmt"
	git "github.com/go-git/go-git/v5"
	http2 "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-github/v32/github"
	"github.com/rotisserie/eris"
	changelogdocutils "github.com/solo-io/go-utils/changeloggenutils"
	. "github.com/solo-io/go-utils/versionutils"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
)

// Default FindDependentVersionFn (used for Gloo Edge)
func FindDependentVersionFn(enterpriseVersion *Version) (*Version, error) {
	versionTag := enterpriseVersion.String()
	dependencyUrl := fmt.Sprintf("https://storage.googleapis.com/gloo-ee-dependencies/%s/dependencies", versionTag[1:])
	request, err := http.NewRequest("GET", dependencyUrl, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(`.*gloo.*(v.*)`)
	if err != nil {
		return nil, err
	}
	matches := re.FindStringSubmatch(string(body))
	if len(matches) != 2 {
		return nil, eris.Errorf("unable to get gloo dependency for gloo enterprise version %s\n response from google storage API: %s", versionTag, string(body))
	}
	glooVersionTag := matches[1]
	version, err := ParseVersion(glooVersionTag)
	if err != nil {
		return nil, err
	}
	return version, nil
}

func main() {
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:               "https://github.com/solo-io/solo-projects",
		Auth: &http2.BasicAuth{
			Username: "nonEmptyString",
			Password: "6d26265c631a07b9998159ee63defc69dedb2bd6",
		},
	})
	_ = func(v *Version) (*Version, error) {
		CheckError(err)
		tagRef, err := repo.Tag(v.String())
		CheckError(err)
		commit, err := repo.CommitObject(tagRef.Hash())
		CheckError(err)
		gomod, err := commit.File("go.mod")
		CheckError(err)
		content, err := gomod.Contents()
		CheckError(err)
		semverRegex := fmt.Sprintf("%s/%s\\s+(v((([0-9]+)\\.([0-9]+)\\.([0-9]+)(?:-([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?)(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?))",
			"solo-io", "gloo")
		regex := regexp.MustCompile(semverRegex)
		// Find version of open source dependency
		matches := regex.FindStringSubmatch(content)
		return ParseVersion(matches[1])
	}
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
	opts := changelogdocutils.Options{
		RepoOwner:     "solo-io",
		MainRepo:      "solo-projects",
		DependentRepo: "gloo",
	}
	depFn, err := changelogdocutils.GetOSDependencyFunc("solo-io",
		"solo-projects",
		"gloo",
		os.Getenv("GITHUB_TOKEN"))
	gen := changelogdocutils.NewMergedReleaseGeneratorWithDepFn(opts, client, depFn )
	changelog, err := gen.GenerateJSON(ctx)
	if err != nil {
		fmt.Println("error", err.Error())
		os.Exit(1)
	}
	fmt.Println("Generated changelog:\n", changelog)
}

func CheckError( err error){
	if err != nil {
		fmt.Printf("err: %s", err.Error())
		os.Exit(1)
	}
}

func getOSDep(v *Version) (*Version, error) {
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:               "https://github.com/solo-io/solo-projects",
		Auth: &http2.BasicAuth{
			Username: "nonEmptyString",
			Password: "6d26265c631a07b9998159ee63defc69dedb2bd6",
		},
		Progress: os.Stdout,
	})
	version := "v1.7.0"
	fmt.Println("cloned")
	CheckError(err)
	tagRef, err := repo.Tag(version)
	CheckError(err)
	commit, err := repo.CommitObject(tagRef.Hash())
	CheckError(err)
	gomod, err := commit.File("go.mod")
	CheckError(err)
	content, err := gomod.Contents()
	CheckError(err)
	fmt.Println(content)
	semverRegex := fmt.Sprintf("%s/%s\\s+(v((([0-9]+)\\.([0-9]+)\\.([0-9]+)(?:-([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?)(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?))",
		"solo-io", "gloo")
	regex := regexp.MustCompile(semverRegex)
	// Find version of open source dependency
	matches := regex.FindStringSubmatch(content)
	fmt.Printf("%+v", matches[1])
	return nil, nil
}
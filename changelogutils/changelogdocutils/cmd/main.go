package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v32/github"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/changelogutils/changelogdocutils"
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
		NumVersions:   200,
		ProjectName:   "",
		RepoOwner:     "solo-io",
		MainRepo:      "solo-projects",
		DependentRepo: "gloo",
	}
	gen := changelogdocutils.NewMergedReleaseGenerator(opts, client)
	changelog, err := gen.GenerateJSON(ctx)
	if err != nil {
		fmt.Println("error", err.Error())
		os.Exit(1)
	}
	f, err := os.Create("./tmp2.json")
	if err != nil {
		fmt.Println(err.Error())
	}
	f.WriteString(changelog)
}
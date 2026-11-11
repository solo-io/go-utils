package githubutils_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/google/go-github/v32/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/githubutils"
)

var _ = Describe("FindLatestTagIncludingPrereleaseBeforeSha against a fake GitHub API", func() {
	const (
		owner = "solo-io"
		repo  = "testrepo"
		sha   = "abc123"
	)

	var server *httptest.Server

	AfterEach(func() {
		server.Close()
	})

	// serves a single page of releases, all of them ancestors of sha
	serveReleases := func(tags ...string) githubutils.RepoClient {
		mux := http.NewServeMux()
		mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/releases", owner, repo), func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("page") != "1" {
				fmt.Fprint(w, `[]`)
				return
			}
			body := "["
			for i, tag := range tags {
				if i > 0 {
					body += ","
				}
				body += fmt.Sprintf(`{"tag_name": %q}`, tag)
			}
			fmt.Fprint(w, body+"]")
		})
		mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/compare/", owner, repo), func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"status": "ahead"}`)
		})
		server = httptest.NewServer(mux)

		client := github.NewClient(nil)
		baseURL, err := url.Parse(server.URL + "/")
		Expect(err).NotTo(HaveOccurred())
		client.BaseURL = baseURL
		return githubutils.NewRepoClient(client, owner, repo)
	}

	DescribeTable("returns the release tag as written",
		func(releaseTag, expected string) {
			client := serveReleases(releaseTag)
			tag, err := client.FindLatestTagIncludingPrereleaseBeforeSha(context.Background(), sha)
			Expect(err).NotTo(HaveOccurred())
			Expect(tag).To(Equal(expected))
		},
		Entry("unlabelled", "v0.71.4", "v0.71.4"),
		Entry("patch0", "v0.71.4-patch0", "v0.71.4-patch0"),
		Entry("patch1", "v0.71.4-patch1", "v0.71.4-patch1"),
		Entry("label without version", "v0.4.3-beta", "v0.4.3"),
	)
})

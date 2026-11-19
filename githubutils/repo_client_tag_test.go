package githubutils_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

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

	type release struct {
		tag      string
		day      string // created_at date, UTC
		ancestor bool
	}

	// serves releases in the given order, two per page; only ancestors compare as "ahead"
	serveDatedReleases := func(releases ...release) githubutils.RepoClient {
		mux := http.NewServeMux()
		mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/releases", owner, repo), func(w http.ResponseWriter, r *http.Request) {
			var page int
			fmt.Sscan(r.URL.Query().Get("page"), &page)
			var items []string
			for i := (page - 1) * 2; i >= 0 && i < len(releases) && i < page*2; i++ {
				items = append(items, fmt.Sprintf(`{"tag_name": %q, "created_at": "%sT12:00:00Z"}`, releases[i].tag, releases[i].day))
			}
			fmt.Fprint(w, "["+strings.Join(items, ",")+"]")
		})
		mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/compare/", owner, repo), func(w http.ResponseWriter, r *http.Request) {
			base := strings.SplitN(strings.TrimPrefix(r.URL.Path, fmt.Sprintf("/repos/%s/%s/compare/", owner, repo)), "...", 2)[0]
			for _, rel := range releases {
				if rel.tag == base && rel.ancestor {
					fmt.Fprint(w, `{"status": "ahead"}`)
					return
				}
			}
			fmt.Fprint(w, `{"status": "diverged"}`)
		})
		server = httptest.NewServer(mux)

		client := github.NewClient(nil)
		baseURL, err := url.Parse(server.URL + "/")
		Expect(err).NotTo(HaveOccurred())
		client.BaseURL = baseURL
		return githubutils.NewRepoClient(client, owner, repo)
	}

	DescribeTable("picks the newest release from the day of the first match",
		func(expected string, releases ...release) {
			client := serveDatedReleases(releases...)
			tag, err := client.FindLatestTagIncludingPrereleaseBeforeSha(context.Background(), sha)
			Expect(err).NotTo(HaveOccurred())
			Expect(tag).To(Equal(expected))
		},
		Entry("final release listed after its rc", "v2.15.0",
			release{"v2.15.0-rc2", "2026-11-17", true},
			release{"v2.15.0", "2026-11-17", true},
			release{"v2.15.0-rc1", "2026-11-10", true},
		),
		Entry("final release on the next page", "v2.15.0",
			release{"v2.16.0-beta1", "2026-11-17", false},
			release{"v2.15.0-rc2", "2026-11-17", true},
			release{"v2.15.0", "2026-11-17", true},
		),
		Entry("same-day release not on the branch", "v2.15.0-rc2",
			release{"v2.15.0-rc2", "2026-11-17", true},
			release{"v2.15.0", "2026-11-17", false},
		),
		Entry("patch0 alias of its base release", "v1.2.3-patch0",
			release{"v1.2.3-patch0", "2026-11-17", true},
			release{"v1.2.3", "2026-11-17", true},
		),
		Entry("older day with a higher version is not checked", "v2.15.0",
			release{"v2.15.0", "2026-11-17", true},
			release{"v2.16.0-rc1", "2026-11-16", true},
		),
		Entry("different pre-release labels keep API order", "v1.0.0-rc1",
			release{"v1.0.0-rc1", "2026-11-17", true},
			release{"v1.0.0-beta3", "2026-11-17", true},
		),
	)
})

package commands

import (
	"path/filepath"
	"testing"

	"github.com/google/go-github/v32/github"
)

func TestCachedReleasesRoundTrip(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "releases.cache")

	if cachedReleasesFileExists(file) {
		t.Fatalf("expected cache file to not exist yet")
	}

	releases := []*github.RepositoryRelease{
		{TagName: github.String("v1.0.0")},
		{TagName: github.String("v1.1.0")},
	}

	if err := writeCachedReleases(file, releases); err != nil {
		t.Fatalf("writeCachedReleases returned error: %v", err)
	}

	if !cachedReleasesFileExists(file) {
		t.Fatalf("expected cache file to exist after write")
	}

	got := getCachedReleases(file)
	if len(got) != len(releases) {
		t.Fatalf("expected %d releases, got %d", len(releases), len(got))
	}
	for i, r := range got {
		if r.GetTagName() != releases[i].GetTagName() {
			t.Errorf("release %d: expected tag %s, got %s", i, releases[i].GetTagName(), r.GetTagName())
		}
	}
}

func TestGetCachedReleasesMissingFile(t *testing.T) {
	if got := getCachedReleases(filepath.Join(t.TempDir(), "does-not-exist")); got != nil {
		t.Fatalf("expected nil releases for missing file, got %v", got)
	}
}

package githubutils

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"go.uber.org/zap"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	GITHUB_TOKEN = "GITHUB_TOKEN"

	STATUS_SUCCESS = "success"
	STATUS_FAILURE = "failure"
	STATUS_ERROR   = "error"
	STATUS_PENDING = "pending"

	COMMIT_FILE_STATUS_ADDED    = "added"
	COMMIT_FILE_STATUS_MODIFIED = "modified"
	COMMIT_FILE_STATUS_DELETED  = "deleted"

	CONTENT_TYPE_FILE      = "file"
	CONTENT_TYPE_DIRECTORY = "dir"
)

func GetGithubToken() (string, error) {
	token, found := os.LookupEnv(GITHUB_TOKEN)
	if !found {
		return "", errors.Errorf("Could not find %s in environment.", GITHUB_TOKEN)
	}
	return token, nil
}

func GetClient(ctx context.Context) (*github.Client, error) {
	token, err := GetGithubToken()
	if err != nil {
		return nil, err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return client, nil
}

func FindStatus(ctx context.Context, client *github.Client, statusLabel, owner, repo, sha string) (*github.RepoStatus, error) {
	logger := contextutils.LoggerFrom(ctx)
	statues, _, err := client.Repositories.ListStatuses(ctx, owner, repo, sha, nil)
	if err != nil {
		logger.Errorw("can't list statuses", err)
		return nil, err
	}

	var currentStatus *github.RepoStatus
	for _, st := range statues {
		if st.Context == nil {
			continue
		}
		if *st.Context == statusLabel {
			currentStatus = st
			break
		}
	}

	return currentStatus, nil
}

func GetFilesFromGit(ctx context.Context, client *github.Client, owner, repo, ref, path string) ([]*github.RepositoryContent, error) {
	var opts *github.RepositoryContentGetOptions
	if ref != "" && ref != "master" {
		opts = &github.RepositoryContentGetOptions{
			Ref: ref,
		}
	}
	var content []*github.RepositoryContent
	single, list, _, err := client.Repositories.GetContents(ctx, owner, repo, path, opts)
	if err != nil {
		return content, err
	}
	if single != nil {
		content = append(content, single)
	}
	content = list
	return content, nil
}

func GetFilesForChangelogVersion(ctx context.Context, client *github.Client, owner, repo, ref, version string) ([]*github.RepositoryContent, error) {
	path := fmt.Sprintf("changelog/%s", version)
	return GetFilesFromGit(ctx, client, owner, repo, ref, path)
}

func GetRawGitFile(ctx context.Context, client *github.Client, content *github.RepositoryContent, owner, repo, ref string) ([]byte, error) {
	if content.GetType() != "file" {
		return nil, fmt.Errorf("content type must be a single file")
	}
	opts := &github.RepositoryContentGetOptions{
		Ref: ref,
	}
	r, err := client.Repositories.DownloadContents(ctx, owner, repo, content.GetPath(), opts)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	byt, err := ioutil.ReadAll(r)
	return byt, err
}

func FindLatestReleaseTagIncudingPrerelease(ctx context.Context, client *github.Client, owner, repo string) (string, error) {
	releases, _, err := client.Repositories.ListReleases(ctx, owner, repo, &github.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, release := range releases {
		if release.GetDraft() {
			continue
		}
		return release.GetTagName(), nil
	}
	return "", errors.Errorf("Could not find any latest release on the first page of releases")
}

func FindLatestReleaseTag(ctx context.Context, client *github.Client, owner, repo string) (string, error) {
	release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	return *release.TagName, nil
}

func MarkInitialPending(ctx context.Context, client *github.Client, owner, repo, sha, description, label string) (*github.RepoStatus, error) {
	return CreateStatus(ctx, client, owner, repo, sha, description, label, STATUS_PENDING)
}

func MarkSuccess(ctx context.Context, client *github.Client, owner, repo, sha, description, label string) (*github.RepoStatus, error) {
	return CreateStatus(ctx, client, owner, repo, sha, description, label, STATUS_SUCCESS)
}

func MarkFailure(ctx context.Context, client *github.Client, owner, repo, sha, description, label string) (*github.RepoStatus, error) {
	return CreateStatus(ctx, client, owner, repo, sha, description, label, STATUS_FAILURE)
}

func MarkError(ctx context.Context, client *github.Client, owner, repo, sha, description, label string) (*github.RepoStatus, error) {
	return CreateStatus(ctx, client, owner, repo, sha, description, label, STATUS_ERROR)
}

func CreateStatus(ctx context.Context, client *github.Client, owner, repo, sha, description, label, state string) (*github.RepoStatus, error) {
	logger := contextutils.LoggerFrom(ctx)

	status := &github.RepoStatus{
		Context:     &label,
		Description: &description,
		State:       &state,
	}
	logger.Debugf("create %s status", state)

	st, _, err := client.Repositories.CreateStatus(ctx, owner, repo, sha, status)
	if err != nil {
		logger.Errorw("can't create status", zap.String("status", state), zap.Error(err))
		return nil, err
	}
	return st, nil
}

// This function writes directly to a writer, so the user is required to close the writer manually
func DownloadRepoArchive(ctx context.Context, client *github.Client, w io.Writer, owner, repo, sha string) error {
	logger := contextutils.LoggerFrom(ctx)
	opt := &github.RepositoryContentGetOptions{
		Ref: sha,
	}

	archiveURL, _, err := client.Repositories.GetArchiveLink(ctx, owner, repo, github.Tarball, opt)
	if err != nil {
		logger.Errorw("can't get archive", zap.Error(err))
		return err
	}

	err = downloadFile(archiveURL.String(), w)
	if err != nil {
		logger.Errorw("can't download file", zap.Error(err))
		return err
	}
	return nil
}

func downloadFile(url string, w io.Writer) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

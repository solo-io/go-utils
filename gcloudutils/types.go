package gcloudutils

import (
	"strings"

	"google.golang.org/api/cloudbuild/v1"
)

const (
	BranchMaster = "master"

	MissingSourceError = "unable to resolve source"
)

func IsMissingSourceError(err string) bool {
	return strings.Contains(err, MissingSourceError)

}

type BuildStatus string

const (
	StatusUnknown       BuildStatus = "STATUS_UNKNOWN"
	StatusQueued                    = "QUEUED"
	StatusWorking                   = "WORKING"
	StatusSuccess                   = "SUCCESS"
	StatusFailure                   = "FAILURE"
	StatusInternalError             = "INTERNAL_ERROR"
	StatusTimeout                   = "TIMEOUT"
	StatusCancelled                 = "CANCELLED"

	GithubBuildStatusLabel       = "bot-build"
	GithubBuildStatusDescription = "build and test the code"
)

func BuildIsQueued(bs string) bool {
	return bs == StatusQueued
}

func BuildIsSuccess(bs string) bool {
	return bs == StatusSuccess
}
func BuildIsFailure(bs string) bool {
	return bs == StatusFailure
}

func BuildIsDone(bs string) bool {
	switch bs {
	case StatusQueued:
		return false
	case StatusWorking:
		return false
	default:
		return true
	}
}

// Used by slack listener to see if the current branch is a valid branch to respond to
func GetValidTriggeredBranchName(m *cloudbuild.Build) (string, bool) {
	// Source can only contain one of the following: (StorageSource, RepoSource)
	// The other will be nil
	if m.Source.RepoSource == nil {
		return "", false
	}

	// Check that branch is master, and the build was triggered via a build trigger, and not manually
	if m.Source.RepoSource.BranchName != BranchMaster && m.BuildTriggerId == "" {
		return m.Source.RepoSource.BranchName, false
	}
	return m.Source.RepoSource.BranchName, true
}

func GetTargetCommitSh(m *cloudbuild.Build) string {
	var commitSha string
	if m.Source.StorageSource != nil {
		var tags Tags = m.Tags
		commitSha = tags.GetSha()
	}
	if m.Source.RepoSource != nil {
		commitSha = m.Source.RepoSource.CommitSha
	}
	return commitSha
}

// attempts to valid release tag
// returns empty string if none exists
func GetReleaseVersionTag(m *cloudbuild.Build) string {
	var validTag string
	if m.Source.StorageSource != nil {
		var tags Tags = m.Tags
		validTag = tags.GetReleaseTag()
	}
	if m.Source.RepoSource != nil {
		validTag = m.Source.RepoSource.TagName

	}
	return validTag
}

// attempts to return repo name
// returns empty string if none exists
func GetRepoName(m *cloudbuild.Build) string {
	var repoName string
	if m.Source.RepoSource != nil {
		spn := GetRepoNameFromRepoSource(m.Source.RepoSource)
		repoName = spn.Repo
	} else {
		var tags Tags = m.Tags
		repoName = tags.GetRepo()
	}
	return repoName
}

// repo source name start as <source>_<owner>_<repo>
// github_solo-io_solobot
// transforms the above into the sum of it's parts
func GetRepoNameFromRepoSource(rs *cloudbuild.RepoSource) *SplitRepoName {
	if rs == nil {
		return nil
	}
	splitName := strings.Split(rs.RepoName, "_")
	if len(splitName) != 3 {
		return nil
	}
	return &SplitRepoName{
		Host: splitName[0],
		Org:  splitName[1],
		Repo: splitName[2],
	}
}

func GetSplitRepoName(m *cloudbuild.Build) *SplitRepoName {
	splitRepoName := &SplitRepoName{
		Org:  "solo-io",
		Host: "github",
	}
	if m.Source.RepoSource != nil {
		spn := GetRepoNameFromRepoSource(m.Source.RepoSource)
		splitRepoName = spn
	} else {
		var tags Tags = m.Tags
		splitRepoName.Repo = tags.GetRepo()
	}
	return splitRepoName
}

type SplitRepoName struct {
	Host string
	Org  string
	Repo string
}

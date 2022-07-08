package securityscanutils

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v32/github"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/githubutils"
)

type GithubRepo struct {
	RepoName string
	Owner    string
}

func (r GithubRepo) Address() string {
	return fmt.Sprintf("github.com/%s/%s", r.Owner, r.RepoName)
}

type GithubIssueWriter struct {
	// The details about the Github repository
	repo GithubRepo

	// The client used to write the issues
	client *github.Client

	// The RepositoryReleasePredicate used to determine if a vulnerability
	// associated with a certain release should be uploaded to GitHub
	createGithubIssuePredicate githubutils.RepositoryReleasePredicate

	// A local cache of all existing GitHub issues
	// Used to ensure that we are updating existing issues that were created by previous scans
	allGithubIssues []*github.Issue
}

func NewGithubIssueWriter(repo GithubRepo, client *github.Client, issuePredicate githubutils.RepositoryReleasePredicate) *GithubIssueWriter {
	return &GithubIssueWriter{
		repo:                       repo,
		client:                     client,
		createGithubIssuePredicate: issuePredicate,
		allGithubIssues:            nil, // initially nil, we'll lazy load these
	}
}

// Labels that are applied to github issues that security scan generates
var labels = []string{"trivy", "vulnerability"}

func (g *GithubIssueWriter) getAllGithubIssues(ctx context.Context) ([]*github.Issue, error) {
	if g.allGithubIssues != nil {
		return g.allGithubIssues, nil
	}

	issues, err := githubutils.GetAllIssues(ctx, g.client, g.repo.Owner, g.repo.RepoName, &github.IssueListByRepoOptions{
		State:  "open",
		Labels: labels,
	})
	if err != nil {
		return nil, eris.Wrapf(err, "error fetching all issues from %s", g.repo.Address())
	}
	g.allGithubIssues = issues
	return g.allGithubIssues, nil
}

// Creates/Updates a Github Issue per image
// The github issue will have the markdown table report of the image's vulnerabilities
// example: https://github.com/solo-io/solo-projects/issues/2458
func (g *GithubIssueWriter) CreateUpdateVulnerabilityIssue(ctx context.Context, release *github.RepositoryRelease, vulnerabilityMarkdown string) error {
	if !g.shouldWriteIssue(release) {
		// The GithubIssueWriter can be configured to only write issues for certain releases
		return nil
	}

	// We can swallow the error here, any releases with improper tag names
	// will not be included in the filtered list
	versionToScan, _ := semver.NewVersion(release.GetTagName())

	issueTitle := fmt.Sprintf("Security Alert: %s", versionToScan.String())
	issueRequest := &github.IssueRequest{
		Title:  github.String(issueTitle),
		Body:   github.String(vulnerabilityMarkdown),
		Labels: &labels,
	}
	createNewIssue := true

	issues, err := g.getAllGithubIssues(ctx)
	if err != nil {
		return eris.Wrapf(err, "failed to get all github issues for repo")
	}

	// TODO - We could avoid iterating over the issues by indexing them by Title
	for _, issue := range issues {
		// If issue already exists, update existing issue with new security scan
		if issue.GetTitle() == issueTitle {
			// Only create new issue if issue does not already exist
			createNewIssue = false
			err := githubutils.UpdateIssue(ctx, g.client, g.repo.Owner, g.repo.RepoName, issue.GetNumber(), issueRequest)
			if err != nil {
				return eris.Wrapf(err, "error updating issue with issue request %+v", issueRequest)
			}
			break
		}
	}
	if createNewIssue {
		_, err := githubutils.CreateIssue(ctx, g.client, g.repo.Owner, g.repo.RepoName, issueRequest)
		if err != nil {
			return eris.Wrapf(err, "error creating issue with issue request %+v", issueRequest)
		}
	}
	return nil
}

func (g *GithubIssueWriter) shouldWriteIssue(release *github.RepositoryRelease) bool {
	return g.createGithubIssuePredicate.Apply(release)
}

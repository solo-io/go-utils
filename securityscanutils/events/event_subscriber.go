package events

import (
	"context"
	"fmt"
	"github.com/google/go-github/v32/github"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/log"
)

type EventSubscriber interface {
	HandleEvent(event Event)
}

type LoggingEventSubscriber struct {
}

func (l *LoggingEventSubscriber) HandleEvent(event Event) {
	log.Printf("Handle Event: %v", event)
}

type GithubSubscription struct {
	// Creates github issue if image vulnerabilities are found
	CreateGithubIssuePerVersion bool
}

func NewGitHubIssueEventSubscriber(config *GithubSubscription) *GitHubIssueEventSubscriber {
	return &GitHubIssueEventSubscriber{
		config: config,
		githubIssuesPerRepository: make(map[string][]*github.Issue),
	}
}

// Creates/Updates a Github Issue per image
// The github issue will have the markdown table report of the image's vulnerabilities
// example: https://github.com/solo-io/solo-projects/issues/2458
type GitHubIssueEventSubscriber struct {
	config *GithubSubscription

	githubClient *github.Client
	githubIssuesPerRepository map[string][]*github.Issue
}

// Labels that are applied to github issues that security scan generates
var TrivyLabels = []string{
	"trivy",
	"vulnerability",
}

func (g *GitHubIssueEventSubscriber) HandleEvent(event Event) {
	ctx := context.Background()

	if !g.config.CreateGithubIssuePerVersion {
		// If we do not need to create a github issue, ignore event
		return
	}

	if event.Topic != VulnerabilityFound {
		return
	}

	vulnerabilityFoundData, ok := event.Data.(VulnerabilityFoundEventData)
	if !ok {
		// invalid event data supplied
		return
	}

	if vulnerabilityFoundData.VulnerabilityMd == "" {
		// don't create an issue for empty markdown
		return
	}

	// get all existing github issues
	githubIssues := g.geAllGithubIssuesForRepo(ctx, vulnerabilityFoundData)

	if err := g.createOrUpdateVulnerabilityIssue(ctx, githubIssues, vulnerabilityFoundData); err != nil {
		contextutils.LoggerFrom(ctx).Errorf("Failed to create or update vulnerability github issue: %v", err)
	}
}

func (g *GitHubIssueEventSubscriber) geAllGithubIssuesForRepo(ctx context.Context, data VulnerabilityFoundEventData) []*github.Issue {
	issues, ok := g.githubIssuesPerRepository[data.RepositoryName]
	if ok {
		return issues
	}

	openIssues, err := githubutils.GetAllIssues(ctx, g.githubClient, data.RepositoryOwner, data.RepositoryName, &github.IssueListByRepoOptions{
		State:  "open",
		Labels: TrivyLabels,
	})

	if err != nil {
		contextutils.LoggerFrom(ctx).Errorf("Failed to fetch issues: %v", err)
	} else {
		g.githubIssuesPerRepository[data.RepositoryName] = openIssues
	}
	return openIssues
}

func (g *GitHubIssueEventSubscriber) createOrUpdateVulnerabilityIssue(ctx context.Context, githubIssues []*github.Issue, data VulnerabilityFoundEventData) error {
	issueTitle := fmt.Sprintf("Security Alert: %s", data.Version)
	issueRequest := &github.IssueRequest{
		Title:  github.String(issueTitle),
		Body:   github.String(data.VulnerabilityMd),
		Labels: &TrivyLabels,
	}
	createNewIssue := true

	for _, issue := range githubIssues {
		// If issue already exists, update existing issue with new security scan
		if issue.GetTitle() == issueTitle {
			// Only create new issue if issue does not already exist
			createNewIssue = false
			err := githubutils.UpdateIssue(ctx, g.githubClient, data.RepositoryOwner, data.RepositoryName, issue.GetNumber(), issueRequest)
			if err != nil {
				return eris.Wrapf(err, "error updating issue with issue request %+v", issueRequest)
			}
			break
		}
	}
	if createNewIssue {
		_, err := githubutils.CreateIssue(ctx, g.githubClient, data.RepositoryOwner, data.RepositoryName, issueRequest)
		if err != nil {
			return eris.Wrapf(err, "error creating issue with issue request %+v", issueRequest)
		}
	}
	return nil
}



type SlackSubscription struct {
	WebhookUrl string
}

func NewSlackNotificationEventSubscriber(config *SlackSubscription) *SlackNotificationEventSubscriber {
	return &SlackNotificationEventSubscriber{
		config: config,
	}
}

type SlackNotificationEventSubscriber struct {
	config *SlackSubscription
}


func (s *SlackNotificationEventSubscriber) HandleEvent(event Event) {
	if s.config.WebhookUrl == "" {
		// No webhookUrl, ignore event
		return
	}
}


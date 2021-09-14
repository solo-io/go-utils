package events

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/slackutils"
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
		config:                    config,
		githubIssuesPerRepository: make(map[string][]*github.Issue),
	}
}

// Creates/Updates a Github Issue per image
// The github issue will have the markdown table report of the image's vulnerabilities
// example: https://github.com/solo-io/solo-projects/issues/2458
type GitHubIssueEventSubscriber struct {
	config *GithubSubscription

	githubClient              *github.Client
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

	// if unable to initialize client, return
	if err := g.setGithubClient(ctx); err != nil {
		contextutils.LoggerFrom(ctx).Errorf("Failed to initialize github client: %v", err)
		return
	}

	// get all existing github issues
	githubIssues := g.geAllGithubIssuesForRepo(ctx, vulnerabilityFoundData)

	if err := g.createOrUpdateVulnerabilityIssue(ctx, githubIssues, vulnerabilityFoundData); err != nil {
		contextutils.LoggerFrom(ctx).Errorf("Failed to create or update vulnerability github issue: %v", err)
	}
}

func (g *GitHubIssueEventSubscriber) setGithubClient(ctx context.Context) error {
	if g.githubClient != nil {
		return nil
	}

	client, err := githubutils.GetClient(ctx)
	if client != nil {
		g.githubClient = client
	}
	return err
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

type CompletedScanMessage struct {
	StartTime time.Time
	EndTime   time.Time

	Result                      string
	VulnerabilitiesFoundPerRepo map[string]int
}

func NewSlackNotificationEventSubscriber(config *SlackSubscription) *SlackNotificationEventSubscriber {
	return &SlackNotificationEventSubscriber{
		config: config,
	}
}

type SlackNotificationEventSubscriber struct {
	config      *SlackSubscription
	slackClient *slackutils.SlackClient

	// aggregate result that will be sent to slack channel
	completedScanMessage *CompletedScanMessage
}

func (s *SlackNotificationEventSubscriber) HandleEvent(event Event) {
	if s.config.WebhookUrl == "" {
		// No webhookUrl, ignore event
		return
	}

	if event.Topic == RepoScanStarted {
		data := event.Data.(EventData)
		// begin new message
		s.completedScanMessage = &CompletedScanMessage{
			StartTime:                   data.Time,
			VulnerabilitiesFoundPerRepo: make(map[string]int),
		}
	}

	if event.Topic == RepoScanCompleted {
		data := event.Data.(EventData)
		// complete the message and flush it
		s.completedScanMessage.EndTime = data.Time
		if data.Err == nil {
			s.completedScanMessage.Result = "SUCCESS"
		} else {
			s.completedScanMessage.Result = fmt.Sprintf("FAILURE: %v", data.Err.Error())
		}
		s.publishCompletedScanMessage(s.completedScanMessage)
	}

	if event.Topic == VulnerabilityFound {
		data := event.Data.(VulnerabilityFoundEventData)

		// increment count
		vulnerabilities, ok := s.completedScanMessage.VulnerabilitiesFoundPerRepo[data.RepositoryName]
		if !ok {
			s.completedScanMessage.VulnerabilitiesFoundPerRepo[data.RepositoryName] = 1
		} else {
			s.completedScanMessage.VulnerabilitiesFoundPerRepo[data.RepositoryName] = vulnerabilities + 1
		}
	}
}

func (s *SlackNotificationEventSubscriber) publishCompletedScanMessage(completedScanMessage *CompletedScanMessage) {
	ctx := context.Background()
	slackClient := slackutils.NewSlackClient(&slackutils.SlackNotifications{
		DefaultUrl: s.config.WebhookUrl,
		RepoUrls:   nil,
	})

	var sb strings.Builder
	sb.WriteString("Security Scan Completed \n")
	sb.WriteString(fmt.Sprintf("Result: %v\n", completedScanMessage.Result))
	sb.WriteString(fmt.Sprintf("Duration: %v\n", completedScanMessage.EndTime.Sub(completedScanMessage.StartTime)))
	for repo, vulnerabilities := range completedScanMessage.VulnerabilitiesFoundPerRepo {
		sb.WriteString(fmt.Sprintf("Found %d vulnerabilities in: %s\n", vulnerabilities, repo))
	}

	slackClient.Notify(ctx, sb.String())
}

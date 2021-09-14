package events

import (
	"context"
	"time"
)

type SecurityScanSubscriptionConfig struct {
	defaultSubscriptions SecurityScanSubscriptions

	repositorySubscriptions map[string]SecurityScanSubscriptions
}

type SecurityScanSubscriptions struct {
	Github *GithubSubscription
	Slack  *SlackSubscription
}

type SecurityScanEventBus struct {
	// user defined configuration
	subscriptionConfig *SecurityScanSubscriptions

	// eventBus mapping topics to subscriptions
	eventBus *EventBus
}

func NewSecurityScanEventBus() *SecurityScanEventBus {
	return &SecurityScanEventBus{
		subscriptionConfig: nil,
		eventBus:           nil,
	}
}

func (n *SecurityScanEventBus) RegisterSubscriptionConfiguration(ctx context.Context, subscriptionConfig *SecurityScanSubscriptions) {
	n.eventBus = NewEventBus()
	n.subscriptionConfig = subscriptionConfig

	// Always log the event
	loggingEventSubscriber := &LoggingEventSubscriber{}
	n.subscribeEventHandlerToTopics(ctx, loggingEventSubscriber, []EventTopic{
		RepoScanStarted,
		RepoScanCompleted,
		VulnerabilityFound,
	})

	// If Github config is defined, register appropriate subscriptions
	if subscriptionConfig.Github != nil {
		githubIssueEventSubscriber := NewGitHubIssueEventSubscriber(subscriptionConfig.Github)
		n.subscribeEventHandlerToTopics(ctx, githubIssueEventSubscriber, []EventTopic{
			VulnerabilityFound,
		})
	}

	// If Slack config is defined, register appropriate subscriptions
	if subscriptionConfig.Slack != nil {
		slackNotificationEventSubscriber := NewSlackNotificationEventSubscriber(subscriptionConfig.Slack)
		n.subscribeEventHandlerToTopics(ctx, slackNotificationEventSubscriber, []EventTopic{
			RepoScanStarted,
			RepoScanCompleted,
			VulnerabilityFound,
		})
	}
}

func (n *SecurityScanEventBus) subscribeEventHandlerToTopics(ctx context.Context, handler EventSubscriber, topics []EventTopic) {
	for _, topic := range topics {
		n.subscribeEventHandlerToTopic(ctx, handler, topic)
	}
}

func (n *SecurityScanEventBus) subscribeEventHandlerToTopic(ctx context.Context, handler EventSubscriber, topic EventTopic) {
	subscriptionChannel := make(EventChannel, 10) // some buffer size to avoid blocking
	go func() {
		for {
			select {
			case event := <-subscriptionChannel:
				handler.HandleEvent(event)
			case <-ctx.Done():
				return
			}
		}
	}()
	n.eventBus.Subscribe(topic, subscriptionChannel)
}

func (n *SecurityScanEventBus) PublishScannerEvent(topic EventTopic, err error) {
	n.publishToTopic(topic, &EventData{
		Time: time.Now(),
		Err:  err,
	})
}

func (n *SecurityScanEventBus) PublishRepositoryEvent(topic EventTopic, repository string, err error) {
	n.publishToTopic(topic, RepositoryEventData{
		EventData: &EventData{
			Time: time.Now(),
			Err:  err,
		},
		RepositoryName: repository,
	})
}

func (n *SecurityScanEventBus) PublishVulnerabilityFound(repositoryName, repositoryOwner, version, vulnerabilityMd string) {
	n.publishToTopic(VulnerabilityFound, &VulnerabilityFoundEventData{
		EventData: &EventData{
			Time: time.Now(),
			Err:  nil,
		},
		RepositoryName:  repositoryName,
		RepositoryOwner: repositoryOwner,
		Version:         version,
		VulnerabilityMd: vulnerabilityMd,
	})
}

func (n *SecurityScanEventBus) publishToTopic(topic EventTopic, data interface{}) {
	n.eventBus.Publish(topic, data)
}

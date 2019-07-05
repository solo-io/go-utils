package gcloudutils

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"golang.org/x/sync/errgroup"
)

type CloudBot struct {
	subscribers []*CloudSubscriber
}

func NewCloudBot(ctx context.Context, projectIds []string, subscriberId string, handlers ...CloudBuildEventHandler) (*CloudBot, error) {
	var subscribers []*CloudSubscriber
	for _, projectId := range projectIds {
		subscriber, err := NewCloudSubscriber(ctx, projectId, subscriberId)
		if err != nil {
			return nil, err
		}
		subscribers = append(subscribers, subscriber)
	}
	for _, subscriber := range subscribers {
		for _, handler := range handlers {
			subscriber.RegisterHandler(handler)
		}
	}
}

func (b *CloudBot) Run(ctx context.Context) error {
	eg := errgroup.Group{}
	for _, subscriber := range b.subscribers {
		sub := subscriber
		eg.Go(func() error {
			subLogger := contextutils.WithLoggerValues(ctx, "projectId", sub.GetProjectId())
			return sub.Run(subLogger)
		})
	}
	return eg.Wait()
}

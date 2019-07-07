package gcloudutils

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type CloudBot struct {
	subscribers []*CloudSubscriber
}

func NewCloudBot(ctx context.Context, projectIds []string, subscriberId string, handlers ...CloudBuildEventHandler) (*CloudBot, error) {
	contextutils.LoggerFrom(ctx).Infow("Creating cloud bot", zap.String("subscriberId", subscriberId))
	var subscribers []*CloudSubscriber
	for _, projectId := range projectIds {
		contextutils.LoggerFrom(ctx).Infow("Adding cloud subscriber for project",
			zap.String("projectId", projectId),
			zap.String("subscriberId", subscriberId))
		subscriber, err := NewCloudSubscriber(ctx, projectId, subscriberId)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("Error creating cloud subscriber",
				zap.Error(err),
				zap.String("projectId", projectId),
				zap.String("subscriberId", subscriberId))
			return nil, err
		}
		subscribers = append(subscribers, subscriber)
	}
	for _, subscriber := range subscribers {
		for _, handler := range handlers {
			subscriber.RegisterHandler(handler)
		}
	}
	return &CloudBot{subscribers: subscribers}, nil
}

func (b *CloudBot) Run(ctx context.Context) error {
	eg := errgroup.Group{}
	for _, subscriber := range b.subscribers {
		sub := subscriber
		eg.Go(func() error {
			subLogger := contextutils.WithLoggerValues(ctx, "projectId", sub.GetProjectId())
			err := sub.Run(subLogger)
			if err != nil {
				contextutils.LoggerFrom(subLogger).Errorw("Error running cloud subscriber", zap.Error(err))
			}
			return err
		})
	}
	return eg.Wait()
}

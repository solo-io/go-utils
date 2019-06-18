package gcloudutils

import (
	"context"
	"encoding/json"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"

	"cloud.google.com/go/pubsub"
	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/githubutils"
	"google.golang.org/api/cloudbuild/v1"
)

const (
	SOLOBOT_SUBSCRIPTION = "solobot"
)

type CloudSubscriber struct {
	githubClient  *github.Client
	buildService  *cloudbuild.Service
	pubSubClient  *pubsub.Client
	cloudBuildSub *pubsub.Subscription
	registry      *Registry
}

func NewSolobotCloudSubscriber(ctx context.Context) (*CloudSubscriber, error) {
	githubClient, err := githubutils.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	buildService, err := NewCloudBuildClient(ctx)
	projectId := GetProjectId()
	pubsubClient, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		return nil, err
	}
	cloudBuildSub := pubsubClient.Subscription(SOLOBOT_SUBSCRIPTION)
	if err != nil {
		return nil, err
	}
	cs := &CloudSubscriber{
		githubClient:  githubClient,
		buildService:  buildService,
		pubSubClient:  pubsubClient,
		cloudBuildSub: cloudBuildSub,
		registry:      &Registry{},
	}
	return cs, nil
}

func (cs *CloudSubscriber) RegisterHandler(handler EventHandler) {
	cs.registry.AddEventHandler(handler)
}

func (cs *CloudSubscriber) Run(ctx context.Context) error {
	sub := cs.pubSubClient.Subscription(cs.cloudBuildSub.ID())
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		msg.Ack()
		contextutils.LoggerFrom(ctx).Debugw("acked message", zap.Any("msg", msg))
		go cs.handleCloudBuildEvent(ctx, msg)
	})
	return err
}

func (cs *CloudSubscriber) handleCloudBuildEvent(ctx context.Context, msg *pubsub.Message) {
	if string(msg.Data) == "" {
		contextutils.LoggerFrom(ctx).Errorw("empty data field found", zap.Any("msg", msg))
		return
	}

	var build cloudbuild.Build
	if err := json.Unmarshal(msg.Data, &build); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error unmarshalling message", zap.Any("msg", msg), zap.Error(err))
		return
	}
	contextutils.LoggerFrom(ctx).Debugw("Build %s %s", build.Id, build.Status)
	// If race condition is found do not even call events, can handle at root
	if err := HandleFailedSourceBuild(ctx, &build); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error handling failed source build", zap.Error(err), zap.Any("build", build))
		return
	}

	for _, eventHandler := range cs.registry.eventHandlers {
		eventHandler := eventHandler
		go func() {
			if err := eventHandler.CloudBuild(ctx, cs.githubClient, &build); err != nil {
				contextutils.LoggerFrom(ctx).Errorw("error handling build", zap.Error(err), zap.Any("build", build))
			}
		}()
	}
}

package gcloudutils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"

	"cloud.google.com/go/pubsub"
	"github.com/solo-io/go-utils/errors"
	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/grpc/status"
)

const (
	TOPIC = "cloud-builds"
)

var (
	alreadyExistsError = func(subscriptionId string) string {
		return fmt.Sprintf("Resource already exists in the project (resource=%s).", subscriptionId)
	}
)

type CloudSubscriber struct {
	buildService  *cloudbuild.Service
	pubsubClient  *pubsub.Client
	cloudBuildSub *pubsub.Subscription
	registry      *CloudBuildRegistry
	projectId     string
}

func NewCloudSubscriber(ctx context.Context, projectId string, subscriptionId string) (*CloudSubscriber, error) {
	buildService, err := NewCloudBuildClient(ctx, projectId)
	contextutils.LoggerFrom(ctx).Debugw("successfully created build service for pubsub", zap.String("projectId", projectId))

	pubsubClient, err := NewPubSubClient(ctx, projectId)
	if err != nil {
		return nil, err
	}
	contextutils.LoggerFrom(ctx).Debugw("Successfully created pubsub client", zap.String("projectId", projectId))

	cloudBuildSub, err := pubsubClient.CreateSubscription(ctx, subscriptionId, pubsub.SubscriptionConfig{
		Topic: pubsubClient.Topic(TOPIC),
	})
	if err != nil {
		if grpcErr, ok := status.FromError(err); ok && grpcErr.Message() != alreadyExistsError(subscriptionId) {
			contextutils.LoggerFrom(ctx).Errorw("Failed on error",
				zap.Error(err),
				zap.String("grpcErr", grpcErr.Message()),
				zap.String("alreadyExistsErr", alreadyExistsError(subscriptionId)))
			return nil, err
		}
		cloudBuildSub = pubsubClient.Subscription(subscriptionId)
	}

	cs := &CloudSubscriber{
		buildService:  buildService,
		pubsubClient:  pubsubClient,
		cloudBuildSub: cloudBuildSub,
		registry:      &CloudBuildRegistry{},
	}
	cs.pubsubClient = pubsubClient
	cs.cloudBuildSub = cloudBuildSub

	contextutils.LoggerFrom(ctx).Debugw("successfully setup pubsub")

	return cs, nil
}

func (cs *CloudSubscriber) RegisterHandler(handler CloudBuildEventHandler) {
	cs.registry.AddEventHandler(handler)
}

func (cs *CloudSubscriber) Run(ctx context.Context) error {
	sub := cs.pubsubClient.Subscription(cs.cloudBuildSub.ID())

	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		msg.Ack()
		go cs.handleCloudBuildEvent(ctx, msg)
	})
	if err != nil && err != context.Canceled {
		return errors.Wrapf(err, "error in subscription for cloud build events")
	}
	return nil
}

func (cs *CloudSubscriber) GetProjectId() string {
	return cs.projectId
}

func (cs *CloudSubscriber) handleCloudBuildEvent(ctx context.Context, msg *pubsub.Message) {
	if string(msg.Data) == "" {
		contextutils.LoggerFrom(ctx).Errorw("empty data field found")
		return
	}

	var cbm cloudbuild.Build
	if err := json.Unmarshal(msg.Data, &cbm); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("unable to wrangle message into cloudbuild message", zap.Error(err))
		return
	}
	contextutils.LoggerFrom(ctx).Debugw("unmarshaled build", zap.Any("build", cbm))

	ctx = contextutils.WithLoggerValues(ctx, zap.String("project-id", cbm.ProjectId), zap.String("build-id", cbm.Id))
	// handle all post release events
	HandleCloudBuildEvent(ctx, cs.registry, &cbm)
}

func HandleCloudBuildEvent(ctx context.Context, registry *CloudBuildRegistry, build *cloudbuild.Build) {
	ctx = contextutils.WithLoggerValues(ctx, zap.String("trigger", "cloud-build"), zap.String("build_id", build.Id))
	for _, eventHandler := range registry.eventHandlers {
		eventHandler := eventHandler
		go func() {
			if err := eventHandler.CloudBuild(ctx, build); err != nil {
				contextutils.LoggerFrom(ctx).Errorw("error handling build", zap.String("build_id", build.Id), zap.Error(err))
			}
		}()
	}
}

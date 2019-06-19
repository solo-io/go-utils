package botutils

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/palantir/go-baseapp/baseapp"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rs/zerolog"
	v1 "github.com/solo-io/build/pkg/api/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/protoutils"
	"goji.io/pat"
)

type StaticBotConfig struct {
	BotName string
}

func Start(ctx context.Context, staticCfg StaticBotConfig, plugins ...Plugin) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}

	cc, err := githubapp.NewDefaultCachingClientCreator(
		cfg.Github,
		githubapp.WithClientUserAgent("changelog-bot/TODO"),
		githubapp.WithClientMiddleware(
			githubapp.ClientLogging(zerolog.DebugLevel),
		),
	)
	if err != nil {
		return err
	}

	var buildCfg v1.BuildConfig
	err = protoutils.UnmarshalYaml([]byte(cfg.AppConfig.BuildConfig), &buildCfg)
	if err != nil {
		return err
	}
	contextutils.LoggerFrom(ctx).Infow(fmt.Sprintf("Hello from %s!", staticCfg.BotName))
	// baseapp library requires zerolog, we use zap everywhere else :(
	logger := zerolog.Ctx(ctx)
	if logger == nil {
		tmp := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &tmp
	}
	server, err := baseapp.NewServer(
		cfg.Server,
		baseapp.DefaultParams(*logger, fmt.Sprintf("%s.", staticCfg.BotName))...,
	)
	if err != nil {
		return err
	}

	githubHandler := NewGithubHookHandler(cc,
		NewConfigFetcher(DefaultRepoCfg, &cfg.AppConfig, &buildCfg))
	for _, p := range plugins {
		githubHandler.RegisterPlugin(p)
	}
	webhookHandler := githubapp.NewDefaultEventDispatcher(cfg.Github, githubHandler)
	server.Mux().Handle(pat.Post(githubapp.DefaultWebhookRoute), webhookHandler)
	server.Mux().Handle(pat.New("/"), _200ok())

	// Start is blocking
	return server.Start()
}

func _200ok() http.Handler {
	return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
}

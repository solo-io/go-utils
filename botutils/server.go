package botutils

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/solo-io/go-utils/botutils/botconfig"

	"github.com/palantir/go-baseapp/baseapp"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rs/zerolog"
	"github.com/solo-io/go-utils/contextutils"
	"goji.io/pat"
)

type StaticBotConfig struct {
	BotName string
	Version string
}

type GitBot interface {
	Start(ctx context.Context, plugins ...Plugin) error
}

// This bot doesn't read any configuration from the repo, just uses the application config and static config provided
type simpleGitBot struct {
	staticConfig StaticBotConfig
	config       *botconfig.Config
}

func NewSimpleGitBot(staticConfig StaticBotConfig) (GitBot, error) {
	config, err := botconfig.ReadConfig()
	if err != nil {
		return nil, err
	}
	return &simpleGitBot{
		config:       config,
		staticConfig: staticConfig,
	}, nil
}

func (b *simpleGitBot) Start(ctx context.Context, plugins ...Plugin) error {
	cc, err := githubapp.NewDefaultCachingClientCreator(
		b.config.Github,
		githubapp.WithClientUserAgent(fmt.Sprintf("%s/%s", b.staticConfig.BotName, b.staticConfig.Version)),
		githubapp.WithClientMiddleware(
			githubapp.ClientLogging(zerolog.DebugLevel),
		),
	)
	if err != nil {
		return err
	}

	contextutils.LoggerFrom(ctx).Infow(fmt.Sprintf("Hello from %s!", b.staticConfig.BotName))
	// baseapp library requires zerolog, we use zap everywhere else :(
	logger := zerolog.Ctx(ctx)
	if logger == nil {
		tmp := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &tmp
	}
	server, err := baseapp.NewServer(
		b.config.Server,
		baseapp.DefaultParams(*logger, fmt.Sprintf("%s.", b.staticConfig.BotName))...,
	)
	if err != nil {
		return err
	}

	githubHandler := NewGithubHookHandler(ctx, cc)
	for _, p := range plugins {
		githubHandler.RegisterPlugin(p)
	}
	webhookHandler := githubapp.NewDefaultEventDispatcher(b.config.Github, githubHandler)
	server.Mux().Handle(pat.Post(githubapp.DefaultWebhookRoute), webhookHandler)
	server.Mux().Handle(pat.New("/debug/pprof/"), http.DefaultServeMux)
	server.Mux().Handle(pat.New("/"), _200ok())

	// Start is blocking
	return server.Start()
}

func _200ok() http.Handler {
	return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
}

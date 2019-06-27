# Botutils

This package contains utilities for writing simple git bot applications. 

## Implementing hooks

Write a plugin that implements one or more of the handler interfaces defined in `interface.go`. 

## Implementing a server

Create an instance of SimpleGitBot:

```go
func Run(ctx context.Context) error {
	staticCfg := botutils.StaticBotConfig{BotName: BotName, Version: version.Version}
	bot := botutils.NewSimpleGitBot(staticCfg)
	return bot.Start(context.TODO(), plugins...)
}
```

## Deploying the bot

The bot needs to be deployed with a config that can be deserialized into the `botconfig.Config` struct. By default, 
this should be available at `/etc/solo-github-app/config.yml`, but can be mounted to a custom location 
by setting the `BOT_CONFIG` environment variable. 
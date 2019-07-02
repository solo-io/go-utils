package botconfig

import (
	"github.com/solo-io/go-utils/osutils"
	"strconv"

	"github.com/palantir/go-baseapp/baseapp"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/palantir/go-githubapp/githubapp"
)

//go:generate mockgen -destination os_mock_test.go -self_package github.com/solo-io/go-utils/botutils/botconfig -package botconfig_test github.com/solo-io/go-utils/osutils OsClient

const (
	DefaultBotCfg  = "/etc/solo-github-app/config.yml"

	BotConfigEnvVar = "BOT_CONFIG"
	WebhookSecretEnvVar = "WEBHOOK_SECRET"
	PrivateKeyEnvVar = "PRIVATE_KEY_FILE"
	IntegrationIdEnvVar = "INTEGRATION_ID"
	InstallationIdEnvVar = "INSTALLATION_ID"
)

var (
	FailedToReadFileError = func(err error, path string) error {
		return errors.Wrapf(err, "failed reading file: %s", path)
	}

	FailedToParseConfigError = func(err error, path string) error {
		return errors.Wrapf(err, "failed parsing configuration file: %s", path)
	}

	FailedToParseEnvVarError = func(err error, name, value string) error {
		return errors.Wrapf(err, "error parsing %s environment variable value: %s", name, value)
	}

	MissingBotConfigValueError = func(name string) error {
		return errors.Errorf("missing important bot config, use %s to provide it or include in config map", name)
	}
)

type Config struct {
	Server baseapp.HTTPConfig `yaml:"server"`
	Github githubapp.Config   `yaml:"github"`

	AppConfig ApplicationConfig `yaml:"app_configuration" json:"appConfiguration"`
}

type SlackNotifications struct {
	DefaultUrl string            `yaml:"default_url" json:"defaultUrl"`
	RepoUrls   map[string]string `yaml:"repo_urls" json:"repoUrls"`
}

type ApplicationConfig struct {
	InstallationId     int                `yaml:"installation_id" json:"installationId"`
	SlackNotifications SlackNotifications `yaml:"slack_notifications" json:"slackNotifications"`
}

func ReadConfig() (*Config, error) {
	configReader := &configReader{
		os: osutils.NewOsClient(),
	}
	return configReader.ReadConfig()
}

type ConfigReader interface {
	ReadConfig() (*Config, error)
}

func NewConfigReader(os osutils.OsClient) ConfigReader {
	return &configReader{
		os: os,
	}
}

type configReader struct {
	os osutils.OsClient
}

// Returns a config based on reading a mounted file containing yaml.
// Rather than including in the config, certain values can be provided through the environment by
// specifying INSTALLATION_ID, INTEGRATION_ID, WEBHOOK_SECRET, and PRIVATE_KEY_FILE.
// The default config location can be overridden with BOT_CONFIG.
// If the config can't be read, parsed, or is missing critical values for the github connection, an
// error will be returned. 
func (r *configReader) ReadConfig() (*Config, error) {
	path := r.os.Getenv(BotConfigEnvVar)
	if path == "" {
		path = DefaultBotCfg
	}

	var c Config

	bytes, err := r.os.ReadFile(path)
	if err != nil {
		return nil, FailedToReadFileError(err, path)
	}

	if err := yaml.UnmarshalStrict(bytes, &c); err != nil {
		return nil, FailedToParseConfigError(err, path)
	}

	r.updateWebhookSecret(&c)
	if err := r.updatePrivateKey(&c); err != nil {
		return nil, err
	}
	if err := r.updateIntegrationId(&c); err != nil {
		return nil, err
	}
	if err := r.updateInstallationId(&c); err != nil {
		return nil, err
	}
	if err := validateConfig(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

func validateConfig(config *Config) error {
	if config.AppConfig.InstallationId == 0 {
		return MissingBotConfigValueError(InstallationIdEnvVar)
	}
	if config.Github.App.PrivateKey == "" {
		return MissingBotConfigValueError(PrivateKeyEnvVar)
	}
	if config.Github.App.IntegrationID == 0 {
		return MissingBotConfigValueError(IntegrationIdEnvVar)
	}
	if config.Github.App.WebhookSecret == "" {
		return MissingBotConfigValueError(WebhookSecretEnvVar)
	}
	return nil
}

func (r *configReader) updateWebhookSecret(config *Config) {
	webhookSecret := r.os.Getenv(WebhookSecretEnvVar)
	if webhookSecret != "" {
		config.Github.App.WebhookSecret = webhookSecret
	}
}

func (r *configReader) updatePrivateKey(config *Config) error {
	privateKeyFile := r.os.Getenv(PrivateKeyEnvVar)
	if privateKeyFile == "" {
		return nil
	}

	bytes, err := r.os.ReadFile(privateKeyFile)
	if err != nil {
		return FailedToReadFileError(err, privateKeyFile)
	}
	config.Github.App.PrivateKey = string(bytes)
	return nil
}

func (r *configReader) updateIntegrationId(config *Config) error {
	integrationIdStr := r.os.Getenv(IntegrationIdEnvVar)
	if integrationIdStr != "" {
		integrationId, err := strconv.Atoi(integrationIdStr)
		if err != nil {
			return FailedToParseEnvVarError(err, IntegrationIdEnvVar, integrationIdStr)
		}
		config.Github.App.IntegrationID = integrationId
	}
	return nil
}

func (r *configReader) updateInstallationId(config *Config) error {
	installationIdStr := r.os.Getenv(InstallationIdEnvVar)
	if installationIdStr != "" {
		installationId, err := strconv.Atoi(installationIdStr)
		if err != nil {
			return FailedToParseEnvVarError(err, InstallationIdEnvVar, installationIdStr)
		}
		config.AppConfig.InstallationId = installationId
	}
	return nil
}
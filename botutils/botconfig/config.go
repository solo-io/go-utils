package botconfig

import (
	"io/ioutil"
	"os"

	"github.com/palantir/go-baseapp/baseapp"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/palantir/go-githubapp/githubapp"
)

const (
	DefaultBotCfg  = "/etc/solo-github-app/config.yml"
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
	GcloudProjects     []string           `yaml:"gcloud_projects" json:"gcloudProjects"`
	SlackNotifications SlackNotifications `yaml:"slack_notifications" json:"slackNotifications"`
	BuildConfig        string             `yaml:"build_config" json:"buildConfig"`
}

func ReadConfig() (*Config, error) {
	path := os.Getenv("BOT_CONFIG")
	if path == "" {
		path = DefaultBotCfg
	}

	var c Config

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed reading server config file: %s", path)
	}

	if err := yaml.UnmarshalStrict(bytes, &c); err != nil {
		return nil, errors.Wrap(err, "failed parsing configuration file")
	}

	return &c, nil
}

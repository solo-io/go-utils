package botconfig_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/palantir/go-baseapp/baseapp"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/botutils/botconfig"
)

var _ = Describe("BotconfigTest", func() {

	var (
		os        *MockOsClient
		reader    botconfig.ConfigReader
		nestedErr = eris.Errorf("")
		ctrl      *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(test)
		os = NewMockOsClient(ctrl)
		reader = botconfig.NewConfigReader(os)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("fails if default bot config can't be read", func() {
		Expect(1).To(Equal(3))
		os.EXPECT().Getenv(botconfig.BotConfigEnvVar).Return("")
		os.EXPECT().ReadFile(botconfig.DefaultBotCfg).Return(nil, nestedErr)

		expected := botconfig.FailedToReadFileError(nestedErr, botconfig.DefaultBotCfg)
		conf, err := reader.ReadConfig()
		Expect(conf).To(BeNil())
		Expect(err.Error()).To(Equal(expected.Error()))
	})

	It("fails if custom bot config can't be read", func() {
		os.EXPECT().Getenv(botconfig.BotConfigEnvVar).Return("foo")
		os.EXPECT().ReadFile("foo").Return(nil, nestedErr)

		expected := botconfig.FailedToReadFileError(nestedErr, "foo")
		conf, err := reader.ReadConfig()
		Expect(conf).To(BeNil())
		Expect(err.Error()).To(Equal(expected.Error()))
	})

	It("fails if bot config can't be parsed", func() {
		os.EXPECT().Getenv(botconfig.BotConfigEnvVar).Return("")
		os.EXPECT().ReadFile(botconfig.DefaultBotCfg).Return([]byte("invalid-yaml"), nil)

		expected := botconfig.FailedToParseConfigError(nestedErr, botconfig.DefaultBotCfg)
		conf, err := reader.ReadConfig()
		Expect(conf).To(BeNil())
		Expect(err.Error()).To(ContainSubstring(expected.Error()))
	})

	It("works if bot config doesn't need overriding", func() {
		os.EXPECT().Getenv(botconfig.BotConfigEnvVar).Return("")
		os.EXPECT().Getenv(botconfig.WebhookSecretEnvVar).Return("")
		os.EXPECT().Getenv(botconfig.PrivateKeyEnvVar).Return("")
		os.EXPECT().Getenv(botconfig.IntegrationIdEnvVar).Return("")
		os.EXPECT().ReadFile(botconfig.DefaultBotCfg).Return([]byte(validConfig), nil)

		conf, err := reader.ReadConfig()
		Expect(err).To(BeNil())
		Expect(conf).To(Equal(getValidConfig()))
	})

	It("works if bot config is fully overridden", func() {
		os.EXPECT().Getenv(botconfig.BotConfigEnvVar).Return("")
		os.EXPECT().Getenv(botconfig.WebhookSecretEnvVar).Return("foo")
		os.EXPECT().Getenv(botconfig.PrivateKeyEnvVar).Return("private-key.file")
		os.EXPECT().Getenv(botconfig.IntegrationIdEnvVar).Return("12345")
		os.EXPECT().ReadFile(botconfig.DefaultBotCfg).Return([]byte(validConfigTrimmed), nil)
		os.EXPECT().ReadFile("private-key.file").Return([]byte("bar\nbaz\n"), nil)

		conf, err := reader.ReadConfig()
		Expect(err).To(BeNil())
		Expect(conf).To(Equal(getValidConfig()))
	})

	It("errors if integration id can't be parsed", func() {
		os.EXPECT().Getenv(botconfig.BotConfigEnvVar).Return("")
		os.EXPECT().Getenv(botconfig.WebhookSecretEnvVar).Return("foo")
		os.EXPECT().Getenv(botconfig.PrivateKeyEnvVar).Return("private-key.file")
		os.EXPECT().Getenv(botconfig.IntegrationIdEnvVar).Return("unparseable")
		os.EXPECT().ReadFile(botconfig.DefaultBotCfg).Return([]byte(validConfigTrimmed), nil)
		os.EXPECT().ReadFile("private-key.file").Return([]byte("bar\nbaz\n"), nil)

		expected := botconfig.FailedToParseEnvVarError(nestedErr, botconfig.IntegrationIdEnvVar, "unparseable")
		conf, err := reader.ReadConfig()
		Expect(conf).To(BeNil())
		Expect(err.Error()).To(ContainSubstring(expected.Error()))
	})

	It("errors if integration id is missing", func() {
		os.EXPECT().Getenv(botconfig.BotConfigEnvVar).Return("")
		os.EXPECT().Getenv(botconfig.WebhookSecretEnvVar).Return("foo")
		os.EXPECT().Getenv(botconfig.PrivateKeyEnvVar).Return("private-key.file")
		os.EXPECT().Getenv(botconfig.IntegrationIdEnvVar).Return("")
		os.EXPECT().ReadFile(botconfig.DefaultBotCfg).Return([]byte(validConfigTrimmed), nil)
		os.EXPECT().ReadFile("private-key.file").Return([]byte("bar\nbaz\n"), nil)

		expected := botconfig.MissingBotConfigValueError(botconfig.IntegrationIdEnvVar)
		conf, err := reader.ReadConfig()
		Expect(conf).To(BeNil())
		Expect(err.Error()).To(ContainSubstring(expected.Error()))
	})

	It("errors if webhook secret is missing", func() {
		os.EXPECT().Getenv(botconfig.BotConfigEnvVar).Return("")
		os.EXPECT().Getenv(botconfig.WebhookSecretEnvVar).Return("")
		os.EXPECT().Getenv(botconfig.PrivateKeyEnvVar).Return("private-key.file")
		os.EXPECT().Getenv(botconfig.IntegrationIdEnvVar).Return("12345")
		os.EXPECT().ReadFile(botconfig.DefaultBotCfg).Return([]byte(validConfigTrimmed), nil)
		os.EXPECT().ReadFile("private-key.file").Return([]byte("bar\nbaz\n"), nil)

		expected := botconfig.MissingBotConfigValueError(botconfig.WebhookSecretEnvVar)
		conf, err := reader.ReadConfig()
		Expect(conf).To(BeNil())
		Expect(err.Error()).To(ContainSubstring(expected.Error()))
	})

	It("errors if private key is missing", func() {
		os.EXPECT().Getenv(botconfig.BotConfigEnvVar).Return("")
		os.EXPECT().Getenv(botconfig.WebhookSecretEnvVar).Return("foo")
		os.EXPECT().Getenv(botconfig.PrivateKeyEnvVar).Return("")
		os.EXPECT().Getenv(botconfig.IntegrationIdEnvVar).Return("12345")
		os.EXPECT().ReadFile(botconfig.DefaultBotCfg).Return([]byte(validConfigTrimmed), nil)

		expected := botconfig.MissingBotConfigValueError(botconfig.PrivateKeyEnvVar)
		conf, err := reader.ReadConfig()
		Expect(conf).To(BeNil())
		Expect(err.Error()).To(ContainSubstring(expected.Error()))
	})

})

func getValidConfig() *botconfig.Config {
	c := getValidConfigTrimmed()
	c.Github.App.IntegrationID = 12345
	c.Github.App.WebhookSecret = "foo"
	c.Github.App.PrivateKey = "bar\nbaz\n"
	return c
}

func getValidConfigTrimmed() *botconfig.Config {
	c := &botconfig.Config{
		Server: baseapp.HTTPConfig{
			Address:   "0.0.0.0",
			Port:      8080,
			PublicURL: "https://fake.url",
		},
		Github: githubapp.Config{
			V3APIURL: "https://api.github.com/",
		},
	}
	c.Github.App.IntegrationID = 12345
	c.Github.App.WebhookSecret = "foo"
	c.Github.App.PrivateKey = "bar\nbaz\n"
	return c
}

const (
	validConfig = `
server:
  address: "0.0.0.0"
  port: 8080
  public_url: "https://fake.url"
github:
  v3_api_url: "https://api.github.com/"
  app:
    integration_id: 12345
    webhook_secret: "foo"
    private_key: |
      bar
      baz
`

	validConfigTrimmed = `
server:
  address: "0.0.0.0"
  port: 8080
  public_url: "https://fake.url"
github:
  v3_api_url: "https://api.github.com/"
`
)

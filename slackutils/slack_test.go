package slackutils_test

import (
	"context"

	"github.com/solo-io/go-utils/slackutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("slack client utils", func() {

	const (
		url       = "foo"
		otherUrl  = "foo2"
		message   = "bar"
		repo      = "baz"
		otherRepo = "baz2"
	)

	var (
		ctx        = context.Background()
		httpClient MockHttpClient
	)

	BeforeEach(func() {
		httpClient = MockHttpClient{}
	})

	It("can use default url when notifying with no repo", func() {
		slackNotifications := slackutils.SlackNotifications{
			DefaultUrl: url,
		}
		slackClient := slackutils.NewSlackClientForHttpClient(&httpClient, &slackNotifications)
		slackClient.Notify(ctx, message)
		Expect(httpClient.CalledUrl).To(Equal(url))
	})

	It("can use default url when notifying for repo that doesn't have specific channel", func() {
		slackNotifications := slackutils.SlackNotifications{
			DefaultUrl: url,
		}
		slackClient := slackutils.NewSlackClientForHttpClient(&httpClient, &slackNotifications)
		slackClient.NotifyForRepo(ctx, repo, message)
		Expect(httpClient.CalledUrl).To(Equal(url))
	})

	It("doesn't use http client when no urls set", func() {
		slackNotifications := slackutils.SlackNotifications{}
		slackClient := slackutils.NewSlackClientForHttpClient(nil, &slackNotifications)
		slackClient.Notify(ctx, message)
		Expect(httpClient.CalledUrl).To(Equal(""))
		slackClient.NotifyForRepo(ctx, repo, message)
		Expect(httpClient.CalledUrl).To(Equal(""))
	})

	It("doesn't use http client when no default urls set", func() {
		slackNotifications := slackutils.SlackNotifications{
			RepoUrls: map[string]string{
				repo: otherUrl,
			},
		}
		slackClient := slackutils.NewSlackClientForHttpClient(nil, &slackNotifications)
		slackClient.Notify(ctx, message)
		Expect(httpClient.CalledUrl).To(Equal(""))
		slackClient.NotifyForRepo(ctx, otherRepo, message)
		Expect(httpClient.CalledUrl).To(Equal(""))
	})

	It("can use specific url for repo", func() {
		slackNotifications := slackutils.SlackNotifications{
			DefaultUrl: url,
			RepoUrls: map[string]string{
				repo: otherUrl,
			},
		}
		slackClient := slackutils.NewSlackClientForHttpClient(&httpClient, &slackNotifications)
		slackClient.Notify(ctx, message)
		Expect(httpClient.CalledUrl).To(Equal(url))
		slackClient.NotifyForRepo(ctx, repo, message)
		Expect(httpClient.CalledUrl).To(Equal(otherUrl))
		slackClient.NotifyForRepo(ctx, otherRepo, message)
		Expect(httpClient.CalledUrl).To(Equal(url))
	})

})

type MockHttpClient struct {
	CalledUrl string
}

func (c *MockHttpClient) PostJsonContent(ctx context.Context, message, url string) {
	c.CalledUrl = url
}

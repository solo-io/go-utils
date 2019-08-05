package builders

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/githubutils"
	"google.golang.org/api/cloudbuild/v1"
)

type mockBuilderContext struct {
	*sharedContext

	tag string
	sha string
}

func DefaultMockBuilderContext(ctx context.Context) (*mockBuilderContext, error) {
	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	sharedContext := NewSharedContext("solo-public", 123, client, nil, "solo-io", "testrepo")
	return NewMockBuilderContext(sharedContext, "v0.1.9", "4fa0be07c7daaafcfb97ffe021fbfa8f51696622"), nil
}

func NewMockBuilderContext(sharedContext *sharedContext, tag string, sha string) *mockBuilderContext {
	return &mockBuilderContext{sharedContext: sharedContext, tag: tag, sha: sha}
}

func (ctx *mockBuilderContext) Tag() string {
	return ctx.tag
}

func (ctx *mockBuilderContext) Sha() string {
	return ctx.sha
}

func (ctx *mockBuilderContext) Client() *github.Client {
	return ctx.client
}

func (ctx *mockBuilderContext) Service() *cloudbuild.Service {
	panic("not supported by this interface")
}

func (ctx *mockBuilderContext) Owner() string {
	return ctx.owner
}

func (ctx *mockBuilderContext) Repo() string {
	return ctx.repo
}

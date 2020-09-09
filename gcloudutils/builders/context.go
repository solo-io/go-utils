package builders

import (
	"context"
	"fmt"

	"github.com/google/go-github/v32/github"
	"github.com/rs/zerolog"
	"google.golang.org/api/cloudbuild/v1"
)

type BuildContext interface {
	Client() *github.Client
	Service() *cloudbuild.Service
	Owner() string
	Repo() string
	ProjectId() string
	InstallationId() int64
	Builder() string
}

type ShaBuildContext interface {
	BuildContext
	Sha() string
}

type RefBuildContext interface {
	BuildContext
	Ref() string
}

type TagBuildContext interface {
	ShaBuildContext
	Tag() string
}

type CommentBuildContext interface {
	ShaBuildContext
	Body() string
}

type sharedContext struct {
	projectId      string
	installationId int64
	client         *github.Client
	service        *cloudbuild.Service

	owner   string
	repo    string
	builder string
}

func NewSharedContext(projectId string, installationId int64, client *github.Client, service *cloudbuild.Service, owner string, repo string) *sharedContext {
	return &sharedContext{projectId: projectId, installationId: installationId, client: client, service: service, owner: owner, repo: repo}
}

func (ctx *sharedContext) Client() *github.Client {
	return ctx.client
}

func (ctx *sharedContext) Service() *cloudbuild.Service {
	return ctx.service
}

func (ctx *sharedContext) ProjectId() string {
	return ctx.projectId
}

func (ctx *sharedContext) InstallationId() int64 {
	return ctx.installationId
}

func (ctx *sharedContext) Builder() string {
	return ctx.builder
}

func (ctx *sharedContext) SetBuilder(builder string) {
	ctx.builder = builder
}

type PullRequestContext struct {
	*sharedContext
	pr *github.PullRequest
}

var _ RefBuildContext = new(PullRequestContext)

func (ctx *PullRequestContext) PullRequest() *github.PullRequest {
	return ctx.pr
}

func (ctx *PullRequestContext) Owner() string {
	return ctx.owner
}

func (ctx *PullRequestContext) Repo() string {
	return ctx.repo
}

func (ctx *PullRequestContext) Sha() string {
	// Ifg merged use base (master)
	if ctx.PullRequest().GetMerged() {
		return ctx.pr.GetMergeCommitSHA()
	}
	return ctx.pr.GetHead().GetSHA()
}

func (ctx *PullRequestContext) Ref() string {
	return ctx.pr.GetHead().GetRef()
}

func NewPullRequestContext(sharedContext *sharedContext, pr *github.PullRequest) *PullRequestContext {
	return &PullRequestContext{sharedContext: sharedContext, pr: pr}
}

type ReleaseContext struct {
	*sharedContext

	release *github.ReleaseEvent
	sha     string
}

var _ TagBuildContext = new(ReleaseContext)

func (ctx *ReleaseContext) Release() *github.ReleaseEvent {
	return ctx.release
}

func (ctx *ReleaseContext) Owner() string {
	return ctx.release.GetRepo().GetOwner().GetLogin()
}

func (ctx *ReleaseContext) Repo() string {
	return ctx.release.GetRepo().GetName()
}

func (ctx *ReleaseContext) Sha() string {
	return ctx.sha
}

func (ctx *ReleaseContext) Ref() string {
	return ctx.release.GetRepo().GetDefaultBranch()
}

func (ctx *ReleaseContext) Tag() string {
	return ctx.release.GetRelease().GetTagName()
}

func NewReleaseContext(ctx context.Context, sharedContext *sharedContext, release *github.ReleaseEvent) (*ReleaseContext, error) {
	logger := zerolog.Ctx(ctx)

	tagName := release.GetRelease().GetTagName()

	owner, repo, client := sharedContext.owner, sharedContext.repo, sharedContext.Client()

	tag, _, err := client.Git.GetRef(ctx, owner, repo, "tags/"+tagName)
	if err != nil {
		logger.Error().Err(err).Msg(fmt.Sprintf("Unable to list tag %s for %s", tagName, repo))
		return nil, err
	}
	sha := tag.GetObject().GetSHA()
	logger.Info().Msg(fmt.Sprintf("tag %s is sha %s", tagName, sha))

	// could not find the sha.. return the release anyuway as it may work for some use cases.
	return &ReleaseContext{sharedContext: sharedContext, sha: sha, release: release}, nil
}

type CommitCommentContext struct {
	*sharedContext

	commentEvent *github.CommitCommentEvent
}

var _ CommentBuildContext = new(CommitCommentContext)

func NewCommitCommentContext(sharedContext *sharedContext, commit *github.CommitCommentEvent) *CommitCommentContext {
	return &CommitCommentContext{sharedContext: sharedContext, commentEvent: commit}
}

func (ctx *CommitCommentContext) CommentEvent() *github.CommitCommentEvent {
	return ctx.commentEvent
}

func (ctx *CommitCommentContext) Owner() string {
	return ctx.CommentEvent().GetRepo().GetOwner().GetLogin()
}

func (ctx *CommitCommentContext) Repo() string {
	return ctx.CommentEvent().GetRepo().GetName()
}

func (ctx *CommitCommentContext) Sha() string {
	return ctx.CommentEvent().GetComment().GetCommitID()
}

func (ctx *CommitCommentContext) Ref() string {
	return ctx.CommentEvent().GetRepo().GetDefaultBranch()
}

func (ctx *CommitCommentContext) Body() string {
	return ctx.CommentEvent().GetComment().GetBody()
}

type IssueCommentContext struct {
	*sharedContext

	commentEvent *github.IssueCommentEvent
	pullRequest  *github.PullRequest
}

var _ CommentBuildContext = new(IssueCommentContext)

func NewIssueCommentContext(sharedContext *sharedContext, pullRequest *github.PullRequest,
	comment *github.IssueCommentEvent) *IssueCommentContext {
	return &IssueCommentContext{sharedContext: sharedContext, commentEvent: comment, pullRequest: pullRequest}
}

func (ctx *IssueCommentContext) CommentEvent() *github.IssueCommentEvent {
	return ctx.commentEvent
}

func (ctx *IssueCommentContext) PullRequest() *github.PullRequest {
	return ctx.pullRequest
}

func (ctx *IssueCommentContext) Owner() string {
	return ctx.CommentEvent().GetRepo().GetOwner().GetLogin()
}

func (ctx *IssueCommentContext) Repo() string {
	return ctx.CommentEvent().GetRepo().GetName()
}

func (ctx *IssueCommentContext) Sha() string {
	return ctx.PullRequest().GetHead().GetSHA()
}

func (ctx *IssueCommentContext) Ref() string {
	return ctx.PullRequest().GetHead().GetRef()
}

func (ctx *IssueCommentContext) Author() string {
	return ctx.CommentEvent().GetComment().GetUser().GetLogin()
}

func (ctx *IssueCommentContext) Body() string {
	return ctx.CommentEvent().GetComment().GetBody()
}

func (ctx *IssueCommentContext) Number() int {
	return ctx.CommentEvent().GetIssue().GetNumber()
}

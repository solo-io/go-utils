package commitutils

import (
	"context"
	"time"

	"github.com/google/go-github/github"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/vfsutils"
	"go.uber.org/zap"
)

var (
	RefNotSetError           = errors.Errorf("Must initialize with set ref before updating files.")
	RefAlreadySetError       = errors.Errorf("Ref was already set.")
	CouldNotFindFileToRename = func(oldPath string) error {
		return errors.Errorf("Could not find file %s in tree", oldPath)
	}
)

type CommitSpec struct {
	Name    string
	Email   string
	Message string
}

type RefUpdater interface {
	SetRef(ctx context.Context, ref *github.Reference) error
	UpdateFile(ctx context.Context, path string, contentUpdater func(string) string) error
	RenameFile(ctx context.Context, oldPath, newPath string) error
	Commit(ctx context.Context, spec CommitSpec) error
}

type githubRefUpdater struct {
	client *github.Client
	owner  string
	repo   string

	ref           *github.Reference
	code          vfsutils.MountedRepo
	filesToCommit []github.TreeEntry
}

func NewGithubRefUpdater(client *github.Client, owner, repo string) RefUpdater {
	return &githubRefUpdater{
		client: client,
		owner:  owner,
		repo:   repo,
	}
}

func (c *githubRefUpdater) SetRef(ctx context.Context, ref *github.Reference) error {
	if c.ref != nil {
		return RefAlreadySetError
	}
	c.ref = ref
	c.code = vfsutils.NewLazilyMountedRepo(c.client, c.owner, c.repo, ref.Object.GetSHA())
	c.filesToCommit = nil
	return nil
}

func (c *githubRefUpdater) UpdateFile(ctx context.Context, path string, contentUpdater func(string) string) error {
	if c.ref == nil {
		return RefNotSetError
	}
	contents, err := c.code.GetFileContents(ctx, path)
	if err != nil {
		return err
	}
	newContents := contentUpdater(string(contents))
	contextutils.LoggerFrom(ctx).Infow("Committing file",
		zap.String("contents", string(contents)),
		zap.String("newContents", newContents))
	c.filesToCommit = append(c.filesToCommit, github.TreeEntry{Path: github.String(path), Type: github.String("blob"), Content: github.String(newContents), Mode: github.String("100644")})
	return nil
}

func (c *githubRefUpdater) Commit(ctx context.Context, spec CommitSpec) error {
	if c.ref == nil {
		return RefNotSetError
	}
	tree, _, err := c.client.Git.CreateTree(ctx, c.code.GetOwner(), c.code.GetRepo(), *c.ref.Object.SHA, c.filesToCommit)
	if err != nil {
		return err
	}
	// Get the parent commit to attach the commit to.
	parent, _, err := c.client.Repositories.GetCommit(ctx, c.code.GetOwner(), c.code.GetRepo(), *c.ref.Object.SHA)
	if err != nil {
		return err
	}
	// This is not always populated, but is needed.
	parent.Commit.SHA = parent.SHA
	// Create the commit using the tree.
	date := time.Now()
	author := &github.CommitAuthor{
		Date:  &date,
		Name:  github.String(spec.Name),
		Email: github.String(spec.Email),
	}
	commit := &github.Commit{
		Author:  author,
		Message: github.String(spec.Message),
		Tree:    tree,
		Parents: []github.Commit{*parent.Commit},
	}
	newCommit, _, err := c.client.Git.CreateCommit(ctx, c.code.GetOwner(), c.code.GetRepo(), commit)
	if err != nil {
		return err
	}

	c.ref.Object.SHA = newCommit.SHA
	_, _, err = c.client.Git.UpdateRef(ctx, c.code.GetOwner(), c.code.GetRepo(), c.ref, false)
	return err
}

func (c *githubRefUpdater) RenameFile(ctx context.Context, oldPath, newPath string) error {
	if c.ref == nil {
		return RefNotSetError
	}
	// Get the parent commit to attach the commit to.
	parent, _, err := c.client.Repositories.GetCommit(ctx, c.code.GetOwner(), c.code.GetRepo(), *c.ref.Object.SHA)
	if err != nil {
		return err
	}
	// find the parent's tree entry for the old path
	tree, _, err := c.client.Git.GetTree(ctx, c.code.GetOwner(), c.code.GetRepo(), parent.GetSHA(), false)
	if err != nil {
		return err
	}
	var found *github.TreeEntry
	for _, entry := range tree.Entries {
		if entry.GetPath() == oldPath {
			found = &entry
		}
	}
	if found == nil {
		return CouldNotFindFileToRename(oldPath)
	}

	updated := github.TreeEntry{
		Path: github.String(newPath),
		Type: found.Type,
		SHA:  found.SHA,
		Mode: found.Mode,
	}
	c.filesToCommit = append(c.filesToCommit, updated)
	return nil
}

package vfsutils

import (
	"context"
	"os"
	"path/filepath"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

var (
	CodeMountingError = func(err error) error {
		return errors.Wrapf(err, "error mounting code")
	}

	ReadFileError = func(err error, path string) error {
		return errors.Wrapf(err, "error reading file %s", path)
	}

	ListFilesError = func(err error, path string) error {
		return errors.Wrapf(err, "error listing files of %s", path)
	}

	InvalidDefinitionError = func(msg string) error {
		return eris.New(msg)
	}
)

type MountedRepo interface {
	GetOwner() string
	GetRepo() string
	GetSha() string
	GetFileContents(ctx context.Context, path string) ([]byte, error)
	ListFiles(ctx context.Context, path string) ([]os.FileInfo, error)
}

type lazilyMountedRepo struct {
	owner string
	repo  string
	sha   string

	fs           afero.Fs
	repoRootPath string
	client       *github.Client
}

func (r *lazilyMountedRepo) GetOwner() string {
	return r.owner
}

func (r *lazilyMountedRepo) GetRepo() string {
	return r.repo
}

func (r *lazilyMountedRepo) GetSha() string {
	return r.sha
}

func (r *lazilyMountedRepo) ensureCodeMounted(ctx context.Context) error {
	if r.repoRootPath == "" {
		if r.client == nil {
			return InvalidDefinitionError("must provide a github client if not using a local filesystem")
		}
		contextutils.LoggerFrom(ctx).Infow("downloading repo archive",
			zap.String("owner", r.owner),
			zap.String("repo", r.repo),
			zap.String("sha", r.sha))
		dir, err := MountCode(r.fs, ctx, r.client, r.owner, r.repo, r.sha)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("Error mounting github code",
				zap.Error(err),
				zap.String("owner", r.owner),
				zap.String("repo", r.repo),
				zap.String("sha", r.sha))
			return CodeMountingError(err)
		}
		contextutils.LoggerFrom(ctx).Infow("successfully mounted repo archive",
			zap.String("owner", r.owner),
			zap.String("repo", r.repo),
			zap.String("sha", r.sha),
			zap.String("repoRootPath", r.repoRootPath))
		r.repoRootPath = dir
	}
	return nil
}

func (r *lazilyMountedRepo) GetFileContents(ctx context.Context, path string) ([]byte, error) {
	if err := r.ensureCodeMounted(ctx); err != nil {
		return nil, err
	}
	fsPath := filepath.Join(r.repoRootPath, path)
	return getFileContents(ctx, r.fs, fsPath)
}

func getFileContents(ctx context.Context, fs afero.Fs, fsPath string) ([]byte, error) {
	fileContent, err := afero.ReadFile(fs, fsPath)
	if err != nil {
		return nil, ReadFileError(err, fsPath)
	}
	return fileContent, nil
}

func (r *lazilyMountedRepo) ListFiles(ctx context.Context, path string) ([]os.FileInfo, error) {
	if err := r.ensureCodeMounted(ctx); err != nil {
		return nil, err
	}
	fsPath := filepath.Join(r.repoRootPath, path)
	return listFiles(ctx, r.fs, fsPath)
}

func listFiles(ctx context.Context, fs afero.Fs, fsPath string) ([]os.FileInfo, error) {
	children, err := afero.ReadDir(fs, fsPath)
	if err != nil {
		return nil, ListFilesError(err, fsPath)
	}
	return children, nil
}

func NewLazilyMountedRepo(client *github.Client, owner, repo, sha string) MountedRepo {
	return &lazilyMountedRepo{
		owner:  owner,
		repo:   repo,
		sha:    sha,
		fs:     afero.NewMemMapFs(),
		client: client,
	}
}

type localFsRepo struct {
	owner        string
	repo         string
	fs           afero.Fs
	repoRootPath string
}

// Creates a mounted repo for a local filesystem, the code must already be checked out at the correct SHA, which is
// not known from this implementation.
func NewLocalMountedRepoForFs(repoRootPath, owner, repo string) (MountedRepo, error) {
	if repoRootPath == "" {
		return nil, InvalidDefinitionError("must provide a repoRootPath when using a local filesystem")
	}
	fs := afero.NewOsFs()
	return &localFsRepo{
		owner:        owner,
		repo:         repo,
		fs:           fs,
		repoRootPath: repoRootPath,
	}, nil
}

func (l *localFsRepo) GetOwner() string {
	return l.owner
}

func (l *localFsRepo) GetRepo() string {
	return l.repo
}

func (l *localFsRepo) GetSha() string {
	return "" // TODO this is an unfortunate abstraction
}

func (l *localFsRepo) GetFileContents(ctx context.Context, path string) ([]byte, error) {
	fsPath := filepath.Join(l.repoRootPath, path)
	return getFileContents(ctx, l.fs, fsPath)
}

func (l *localFsRepo) ListFiles(ctx context.Context, path string) ([]os.FileInfo, error) {
	fsPath := filepath.Join(l.repoRootPath, path)
	return listFiles(ctx, l.fs, fsPath)
}

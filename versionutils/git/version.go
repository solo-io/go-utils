package git

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/solo-io/go-utils/errors"
)

var (
	FailedCommandError = func(err error, args []string, output string) error {
		return errors.Wrapf(err, "%v failed: %s", args, output)
	}
)

// Contains different representations of a git ref.
// Hash and tag are always set, branch might be empty if the repo is in detached HEAD state.
type RefInfo struct {
	Branch string
	Hash   string
	Tag    string
}

func GetGitRefInfo(relativeRepoDir string) (*RefInfo, error) {
	info := &RefInfo{}
	repo := gitRepo{relativeDir: relativeRepoDir}

	if tag, err := repo.getTag(); err != nil {
		return nil, err
	} else {
		info.Tag = tag
	}

	if hash, err := repo.getCommitHash(); err != nil {
		return nil, err
	} else {
		info.Hash = hash
	}

	if branch, err := repo.getBranch(); err != nil {
		return nil, err
	} else {
		info.Branch = branch
	}

	return info, nil
}

func PinDependencyVersion(relativeRepoDir string, refName string) error {
	cmd := exec.Command("git", "checkout", refName)
	cmd.Dir = relativeRepoDir
	buf := &bytes.Buffer{}
	out := io.MultiWriter(buf, os.Stdout)
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		return FailedCommandError(err, cmd.Args, buf.String())
	}
	return nil
}

func AppendTagPrefix(version string) string {
	if strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

type gitRepo struct {
	relativeDir string
}

func (g gitRepo) getTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--dirty")
	cmd.Dir = g.relativeDir
	output, err := cmd.Output()
	if err != nil {
		return "", FailedCommandError(err, cmd.Args, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

func (g gitRepo) getCommitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = g.relativeDir
	output, err := cmd.Output()
	if err != nil {
		return "", FailedCommandError(err, cmd.Args, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

func (g gitRepo) getBranch() (string, error) {
	cmd := exec.Command("git", "branch")
	cmd.Dir = g.relativeDir
	output, err := cmd.Output()
	if err != nil {
		return "", FailedCommandError(err, cmd.Args, string(output))
	}

	lines := strings.Split(string(output), "\n")

	var currentBranchLine string
	for _, line := range lines {
		// The line referring to the current branch starts with a "*" character
		if strings.HasPrefix(line, "*") {
			currentBranchLine = line
			break
		}
	}

	// If we are in detached head state, the command outputs "* (HEAD detached at <the_sha>)"
	// Please don't include "HEAD" in your branch names.
	if strings.Contains(currentBranchLine, "HEAD") {
		return "", nil
	}

	return strings.Split(currentBranchLine, " ")[1], nil
}

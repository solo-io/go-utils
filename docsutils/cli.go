package docsutils

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/onsi/ginkgo"
	"github.com/solo-io/go-utils/changelogutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/logger"
	"github.com/solo-io/go-utils/versionutils"
	"github.com/spf13/afero"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	DocsRepo = "solo-docs"
)

type DocsPRSpec struct {
	// Repo owner, i.e. "solo-io"
	Owner           string
	// Repo, i.e. "solo-projects"
	Repo            string
	// Release tag, i.e. "v0.8.5"
	Tag             string
	// Product in docs, i.e. "gloo"
	Product         string
	// Prefix to the changelog file in the docs repo, i.e. "glooe" -> push changelog to solo-docs/<product>/docs/changelog/glooe-changelog
	ChangelogPrefix string
	// Path to the directory containing "docs", i.e. "projects/gloo/doc
	DocsParentPath  string
	// Paths to generated API doc directory that should be copied into docs, i.e. "docs/v1/github.com/solo-io/gloo"
	ApiPaths        []string // can be nil
	// Prefix of the CLI docs files, i.e. "glooctl"
	CliPrefix       string   // empty means don't copy docs files
}

func PushDocsCli(spec *DocsPRSpec) {
	version := versionutils.GetReleaseVersionOrExitGracefully()
	spec.Tag = version.String()
	err := CreateDocsPRFromSpec(spec)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

/*
Useful for cases where repo == docs product name == changelogPrefix name
 */
func CreateDocsPRSimple(owner, repo, tag string, paths ...string) error {
	spec := DocsPRSpec{
		Owner: owner,
		Repo: repo,
		Tag: tag,
		Product: repo,
		ChangelogPrefix: repo,
		ApiPaths: paths,
		CliPrefix: "",
		DocsParentPath: "",
	}
	return CreateDocsPRFromSpec(&spec)
}

/*
Example:
CreateDocsPR("solo-io", "gloo", "gloo", "gloo", "v0.8.2",
"docs/v1/github.com/solo-io/gloo",
"docs/v1/github.com/solo-io/solo-kit",
"docs/v1/gogoproto",
"docs/v1/google")
 */
func CreateDocsPR(owner, repo, product, changelogPrefix, tag string, apiPaths ...string) error {
	spec := DocsPRSpec{
		Owner: owner,
		Repo: repo,
		Tag: tag,
		Product: product,
		ChangelogPrefix: changelogPrefix,
		ApiPaths: apiPaths,
		CliPrefix: "",
		DocsParentPath: "",
	}
	return CreateDocsPRFromSpec(&spec)
}

func validateSpec(spec *DocsPRSpec) error {
	if spec.Owner == "" {
		return errors.Errorf("Owner must not be empty")
	}
	if spec.Repo == "" {
		return errors.Errorf("Repo must not be empty")
	}
	if spec.Tag == "" {
		return errors.Errorf("Tag must not be empty")
	}
	if spec.Product == "" {
		return errors.Errorf("Product must not be empty")
	}
	return nil
}

func CreateDocsPRFromSpec(spec *DocsPRSpec) error {
	ctx := context.TODO()
	fs := afero.NewOsFs()

	err := validateSpec(spec)
	if err != nil {
		return err
	}

	exists, err := afero.Exists(fs, DocsRepo)
	if err != nil {
		return err
	}
	if exists {
		return errors.Errorf("Cannot clone because %s already exists", DocsRepo)
	}

	// setup repo
	err = gitCloneDocs()
	defer fs.RemoveAll(DocsRepo)
	if err != nil {
		return errors.Wrapf(err, "Error cloning repo")
	}

	// setup branch
	branch := spec.Repo + "-docs-" + spec.Tag + "-" + randString(4)
	err = gitCheckoutNewBranch(branch)
	if err != nil {
		return errors.Wrapf(err, "Error checking out branch")
	}

	// update changelog if "changelog" directory exists in this repo
	err = updateChangelogIfNecessary(fs, spec.Tag, spec.Product, spec.ChangelogPrefix)
	if err != nil {
		return err
	}

	// replaceDirectories("gloo", "docs/v1") updates replaces contents of "solo-docs/gloo/docs/v1" with what's in "docs/v1"
	err = replaceApiDirectories(spec.Product, spec.DocsParentPath, spec.ApiPaths...)
	if err != nil {
		return errors.Wrapf(err, "Error removing old docs")
	}

	if spec.CliPrefix != "" {
		replaceCliDocs(spec.Product, spec.DocsParentPath, spec.CliPrefix)
	}

	// see if there is something to commit, push and open PR if so
	return submitPRIfChanges(ctx, spec.Owner, branch, spec.Tag, spec.Product)

	return nil
}

func replaceCliDocs(product, cliPrefix, cliPath string) error {
	// replaceCliDocs(gloo, glooctl, projects/gloo/doc/docs/cli) =>
	//   rm solo-docs/gloo/docs/cli/glooctl*
	//   cp projects/gloo/doc/docs/cli/glooctl* solo-docs/gloo/docs/cli/

	cliDocsDir := filepath.Join("docs/cli")
	soloCliDocsDir := filepath.Join(DocsRepo, product, cliDocsDir)
	oldDocs := filepath.Join(soloCliDocsDir, cliPrefix + "*")
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("rm %s", oldDocs))
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "Could not delete old docs %s, %s", oldDocs)
	}

	newDocs := filepath.Join(cliDocsDir, cliPrefix + "*")
	cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("cp %s %s", newDocs, soloCliDocsDir))
	err = cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "Could not copy new docs %s", oldDocs)
	}
	return nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func randString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func submitPRIfChanges(ctx context.Context, owner, branch, tag, product string) error {
	err := gitAddAll()
	if err != nil {
		return errors.Wrapf(err, "Error doing git add")
	}
	empty, err := gitDiffIsEmpty()
	if err != nil {
		return errors.Wrapf(err, "Error checking for diff")
	}
	if empty {
		// no diff, exit early cause we're done
		return nil
	}
	err = gitCommit(tag)
	if err != nil {
		return errors.Wrapf(err, "Error doing git commit")
	}
	// make sure we can get the client before starting to push
	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return err
	}
	err = gitPush(branch)
	if err != nil {
		return errors.Wrapf(err, "Error pushing docs branch")
	}

	title := fmt.Sprintf("Update docs for %s %s", product, tag)
	body := fmt.Sprintf("Automatically generated docs for %s %s", product, tag)
	base := "master"
	pr := github.NewPullRequest{
		Title: &title,
		Body: &body,
		Head: &branch,
		Base: &base,
	}

	_, _, err = client.PullRequests.Create(ctx, owner, DocsRepo, &pr)
	if err != nil {
		return errors.Wrapf(err, "Error creating PR")
	}
	return nil
}

func updateChangelogIfNecessary(fs afero.Fs, tag, product, changelogPrefix string) error {
	exists, err := changelogutils.ChangelogDirExists(fs, "")
	if err != nil {
		return errors.Wrapf(err, "Error checking for changelog dir")
	}
	if exists {
		if changelogPrefix == "" {
			return errors.Errorf("ChangelogPrefix must be set.")
		}
		changelog, err := changelogutils.ComputeChangelogForTag(fs, tag, "")
		if err != nil {
			return err
		}
		markdown := changelogutils.GenerateChangelogMarkdown(changelog)
		fmt.Printf(markdown)
		err = updateChangelogFile(fs, product, changelogPrefix, markdown, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

func getChangelogDir(product string) string {
	return filepath.Join(DocsRepo, product, "docs", "changelog")
}

// getChangelogFile(gloo, glooe) -> solo-docs/gloo/changelog/glooe-changelog
func getChangelogFile(product, changelogPrefix string) string {
	return filepath.Join(getChangelogDir(product), changelogPrefix + "-changelog")
}

// requires changelog dir to be setup, does not require changelog file to exist
func updateChangelogFile(fs afero.Fs, product, changelogPrefix, markdown, tag string) error {
	changelogDir := getChangelogDir(product)
	changelogFile := getChangelogFile(product, changelogPrefix)
	newContents := fmt.Sprintf("### %s\n\n%s", tag, markdown)
	exists, err := afero.Exists(fs, changelogFile)
	if err != nil {
		return err
	}
	if !exists {
		dirExists, err := afero.Exists(fs, changelogDir)
		if err != nil {
			return err
		} else if !dirExists {
			return errors.Errorf("Directory %s must be set up for changelogs first", changelogDir)
		}
	} else {
		bytes, err := afero.ReadFile(fs, changelogFile)
		if err != nil {
			return err
		}
		oldContents := string(bytes)
		if strings.Contains(oldContents, newContents) {
			// re-release, don't do anything
			return nil
		}
		newContents = newContents + string(bytes)
	}
	return afero.WriteFile(fs, changelogFile, []byte(newContents), 0700)
}

func gitCloneDocs() error {
	token, err := githubutils.GetGithubToken()
	if err != nil {
		return err
	}
	return execGit("", "clone", fmt.Sprintf("https://soloio-bot:%s@github.com/solo-io/%s.git", token, DocsRepo))
}

func gitCheckoutNewBranch(branch string) error {
	return execGit(DocsRepo, "checkout", "-b", branch)
}

func gitAddAll() error {
	return execGit(DocsRepo, "add", ".")
}

func gitCommit(tag string) error {
	return execGit(DocsRepo, "commit", "-m", fmt.Sprintf("docs for %s", tag))
}

func gitDiffIsEmpty() (bool, error) {
	output, err := execGitWithOutput(DocsRepo, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return output == "", err
}

func gitPush(branch string) error {
	return execGit(DocsRepo, "push", "origin", branch)
}

func prepareGitCmd(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	logger.Debugf("git %v", cmd.Args)
	cmd.Env = os.Environ()
	// disable DEBUG=1 from getting through to kube
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter
	cmd.Dir = dir
	return cmd
}

func execGit(dir string, args ...string) error {
	cmd := prepareGitCmd(dir, args...)
	return cmd.Run()
}

func execGitWithOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func replaceApiDirectories(product, docsParentPath string, paths ...string) error {
	fs := afero.NewOsFs()
	for _, path := range paths {
		docsPath := filepath.Join(docsParentPath, path)
		soloDocsPath := filepath.Join(DocsRepo, product, path)
		exists, err := afero.Exists(fs, soloDocsPath)
		if err != nil {
			return err
		}
		if exists {
			err = fs.RemoveAll(soloDocsPath)
			if err != nil {
				return err
			}
		} else {
			err = fs.MkdirAll(soloDocsPath, 0700)
			if err != nil {
				return err
			}
		}
		err = copyRecursive(docsPath, soloDocsPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func copyRecursive(from, to string) error {
	cmd := exec.Command("cp", "-r", from, to)
	logger.Debugf("cp %v", cmd.Args)
	cmd.Env = os.Environ()
	// disable DEBUG=1 from getting through to kube
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter
	return cmd.Run()
}

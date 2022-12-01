package commands

import (
    "context"
    "fmt"

    "github.com/Masterminds/semver/v3"
    "github.com/solo-io/go-utils/cliutils"
    "github.com/solo-io/go-utils/securityscanutils"
    "github.com/spf13/cobra"
    "github.com/spf13/pflag"
)

func ScanRepoCommand(ctx context.Context, rootOptions *RootOptions) *cobra.Command {
    opts := &scanRepoOptions{
        RootOptions: rootOptions,
    }

    cmd := &cobra.Command{
        Use:   "scan-repo",
        Short: "Run Trivy scans against images for the repo specified and upload scan results to a google cloud bucket",
        RunE: func(cmd *cobra.Command, args []string) error {
            return doScanRepo(ctx, opts)
        },
    }
    opts.addToFlags(cmd.Flags())

    return cmd
}

type scanRepoOptions struct {
    *RootOptions

    githubRepository string
    imageRepository  string

    // action to take when a vulnerability is discovered. supported actions are:
    //  none (default): do nothing when a vulnerability is discovered
    //  github-issue-latest (preferred): create a github issue only for the latest patch version of each minor version, when a vulnerability is discovered
    //  github-issue-all: create a github issue for every version where a vulnerability is discovered
    vulnerabilityAction string

    releaseVersionConstraint    string
    imagesVersionConstraintFile string
}

func (m *scanRepoOptions) addToFlags(flags *pflag.FlagSet) {
    flags.StringVarP(&m.githubRepository, "github-repo", "g", "", "github repository to scan")
    flags.StringVarP(&m.imageRepository, "image-repo", "r", securityscanutils.QuayRepository, "image repository to scan")

    flags.StringVarP(&m.vulnerabilityAction, "vulnerability-action", "a", "none", "action to take when a vulnerability is discovered {none, github-issue-all, github-issue-latest}")

    flags.StringVarP(&m.releaseVersionConstraint, "release-constraint", "c", "", "version constraint for releases to scan")
    flags.StringVarP(&m.imagesVersionConstraintFile, "image-constraint-file", "i", "", "name of file with mapping of version to images")

    cliutils.MustMarkFlagRequired(flags, "github-repo")
    cliutils.MustMarkFlagRequired(flags, "release-constraint")
    cliutils.MustMarkFlagRequired(flags, "image-constraint-file")
}

func doScanRepo(ctx context.Context, opts *scanRepoOptions) error {
    releaseVersionConstraint, err := semver.NewConstraint(fmt.Sprintf("%s", opts.releaseVersionConstraint))
    if err != nil {
        return err
    }
    imagesPerVersion, err := GetImagesPerVersionFromFile(opts.imagesVersionConstraintFile)
    if err != nil {
        return err
    }

    scanner := &securityscanutils.SecurityScanner{
        Repos: []*securityscanutils.SecurityScanRepo{
            {
                Repo:  opts.githubRepository,
                Owner: securityscanutils.GithubRepositoryOwner,
                Opts: &securityscanutils.SecurityScanOpts{
                    OutputDir:                              securityscanutils.OutputScanDirectory,
                    ImagesPerVersion:                       imagesPerVersion,
                    VersionConstraint:                      releaseVersionConstraint,
                    ImageRepo:                              opts.imageRepository,
                    CreateGithubIssuePerVersion:            opts.vulnerabilityAction == "github-issue-all",
                    CreateGithubIssueForLatestPatchVersion: opts.vulnerabilityAction == "github-issue-latest",
                },
            },
        },
    }
    return scanner.GenerateSecurityScans(ctx)
}

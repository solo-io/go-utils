package internal_test

import (
    "context"

    "github.com/google/go-github/v32/github"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "github.com/solo-io/go-utils/githubutils"
    "github.com/solo-io/go-utils/pkgmgmtutils/formula_updater_types"
    "github.com/solo-io/go-utils/pkgmgmtutils/internal"
)

var _ = Describe("FormulaBytesUpdater", func() {
    It("can replace submatch", func() {
        regex := `version\s*"([0-9.]+)"`
        testData := `Some other data version "1.2.3" and even more data`

        b, err := internal.UpdateFormulaBytes(
            []byte(testData),
            "4.5.6",
            "git-commit-sha",
            &formula_updater_types.PerPlatformSha256{},
            &formula_updater_types.FormulaOptions{
                VersionRegex: regex,
            })

        Expect(err).NotTo(HaveOccurred())
        Expect(b).To(Equal([]byte(`Some other data version "4.5.6" and even more data`)))
    })

    It("can replace submatch fail", func() {
        regex := `version\s*"([a-z.]+)"`
        testData := `Some other data version "1.2.3" and version "4.5.6" and even more data`

        b, err := internal.UpdateFormulaBytes(
            []byte(testData),
            "7.8.9",
            "git-commit-sha",
            &formula_updater_types.PerPlatformSha256{},
            &formula_updater_types.FormulaOptions{VersionRegex: regex},
        )
        Expect(err).NotTo(HaveOccurred())
        Expect(b).To(Equal([]byte(testData)))
    })

    It("can replace submatch fail", func() {
        regex := `version\s*"[0-9.]+"`
        testData := `Some other data version "1.2.3" and version "4.5.6" and even more data`

        Expect(func() {
            _, _ = internal.UpdateFormulaBytes(
                []byte(testData),
                "7.8.9",
                "git-commit-sha",
                &formula_updater_types.PerPlatformSha256{},
                &formula_updater_types.FormulaOptions{VersionRegex: regex},
            )
        }).To(Panic())
    })

    It("can replace homebrew-core version and sha", func() {
        testVersion := "v7.8.9"
        testVersionSha := "gitcommitsha"
        formula := formula_updater_types.FormulaOptions{
            Name:            "homebrew-core/glooctl",
            FormulaName:     "glooctl",
            Path:            "Formula/glooctl.rb",
            RepoOwner:       "solo-io",
            RepoName:        "homebrew-core",
            PRRepoOwner:     "homebrew",
            PRRepoName:      "homebrew-core",
            PRBranch:        "master",
            PRDescription:   "Created by Solo-io Bot",
            PRCommitName:    "Solo-io Bot",
            PRCommitEmail:   "bot@solo.io",
            VersionRegex:    `tag:\s*"v([0-9.]+)",`,
            VersionShaRegex: `revision:\s*"(.*)"`,
        }

        // GitHub API docs: https://developer.github.com/v3/repos/contents/#get-contents
        gclient := github.NewClient(nil)
        ctx := context.Background()
        fileContents, err := githubutils.GetFilesFromGit(ctx, gclient, formula.RepoOwner, formula.RepoName, "refs/heads/master", formula.Path)
        Expect(len(fileContents)).To(Equal(1))
        Expect(err).To(Not(HaveOccurred()))

        c, err := fileContents[0].GetContent()
        Expect(err).To(Not(HaveOccurred()))

        b, err := internal.UpdateFormulaBytes(
            []byte(c),
            "7.8.9",
            "gitcommitsha",
            &formula_updater_types.PerPlatformSha256{},
            &formula_updater_types.FormulaOptions{VersionRegex: formula.VersionRegex, VersionShaRegex: formula.VersionShaRegex},
        )
        Expect(err).NotTo(HaveOccurred())
        Expect(string(b)).To(ContainSubstring(testVersion))
        Expect(string(b)).To(ContainSubstring(testVersionSha))
    })
})

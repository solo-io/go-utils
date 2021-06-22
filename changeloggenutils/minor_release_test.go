package changelogdocutils_test

import (
	"context"

	"github.com/google/go-github/v32/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/changeloggenutils"
	"github.com/solo-io/go-utils/githubutils"
	. "github.com/solo-io/go-utils/versionutils"
)

func openSourceNotes() *string {
	output := `
**Helm Changes**
- Open Source Helm Change

**Fixes**
- Open Source Fix
`
	return &output
}

var openSourceReleases = []*github.RepositoryRelease{
	{
		TagName: getTagName("v1.2.0"),
		Body:    openSourceNotes(),
	},
	{
		TagName: getTagName("v1.2.0-beta12"),
		Body:    openSourceNotes(),
	},
	{
		TagName: getTagName("v1.2.0-rc2"),
		Body:    openSourceNotes(),
	},
	{
		TagName: getTagName("v1.1.0"),
		Body:    openSourceNotes(),
	},
	{
		TagName: getTagName("v1.1.0-beta9"),
		Body:    openSourceNotes(),
	},
}

var _ = Describe("minor release changelog generator", func() {
	var (
		generator *MinorReleaseGroupedChangelogGenerator
		ctx       context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		client, err := githubutils.GetClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		generator = NewMinorReleaseGroupedChangelogGenerator(Options{
			RepoOwner: "solo-io",
			MainRepo:  "gloo",
		}, client)
	})

	Context("Release Data", func() {
		var (
			releaseData *ReleaseData
		)
		BeforeEach(func() {
			var err error
			releaseData, err = generator.NewReleaseData(openSourceReleases)
			Expect(err).NotTo(HaveOccurred())
		})

		It("groups minor versions correctly", func() {

			for _, release := range openSourceReleases {
				releaseVersion, err := ParseVersion(release.GetTagName())
				Expect(err).NotTo(HaveOccurred())

				minorVersionGroup := GetMajorAndMinorVersion(releaseVersion)
				Expect(releaseData.Releases).To(HaveKey(minorVersionGroup))

				changelogNotes := releaseData.Releases[minorVersionGroup].ChangelogNotes
				Expect(changelogNotes).To(HaveKey(*releaseVersion))
				json, err := changelogNotes[*releaseVersion].Dump()
				Expect(err).NotTo(HaveOccurred())
				Expect(json).To(Equal(`{"Categories":{"Fixes":[{"Note":"Open Source Fix"}],"Helm Changes":[{"Note":"Open Source Helm Change"}]},"ExtraNotes":null,"HeaderSuffix":"","CreatedAt":-62135596800}`))
			}
		})
	})

})

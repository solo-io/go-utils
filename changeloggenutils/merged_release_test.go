package changelogdocutils_test

import (
	"context"
	"fmt"
	"github.com/google/go-github/v32/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/changeloggenutils"
	"github.com/solo-io/go-utils/githubutils"
	. "github.com/solo-io/go-utils/versionutils"
)

func getTagName(tag string) *string {
	s := tag
	return &s
}

func enterpriseNotes() *string {
	output := `
**Enterprise Only Changes**
- Enterprise Only Change

**Helm Changes**
- Enterprise Helm Change

**Fixes**
- Enterprise Fix
`
	return &output
}

var _ = Describe("Merged enterprise and open source release enterpriseNotes", func() {
	var (
		generator *MergedReleaseGenerator
		ctx       context.Context
		enterpriseReleases []*github.RepositoryRelease
	)

	// Mock function to provide which open source version an enterprise version depends on (to merge changelog enterpriseNotes)
	depFn := func (v *Version) (*Version, error){
		// Build dependency tree
		deps := map[string]string{
			"v1.2.1": "v1.2.0",
			"v1.2.0-rc1": "v1.2.0-beta12",
			"v1.2.0-beta13": "v1.1.0",
		}
		version, ok := deps[v.String()]
		if !ok {
			return nil, fmt.Errorf("unable to find dependency for version %s", v.String())
		}
		returnVer, err := ParseVersion(version)
		if err != nil {
			return nil, err
		}
		return returnVer, nil
	}

	BeforeEach(func() {
		ctx = context.Background()
		client, err := githubutils.GetClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		opts := Options{
			RepoOwner:     "solo-io",
			MainRepo:      "solo-projects",
			DependentRepo: "gloo",
		}
		generator = NewMergedReleaseGeneratorWithDepFn(opts, client, depFn)

		enterpriseReleases = []*github.RepositoryRelease{
			{
				TagName: getTagName("v1.2.1"),
				Body:    enterpriseNotes(),
			},
			{
				TagName: getTagName("v1.2.0-beta13"),
				Body:    enterpriseNotes(),
			},
			{
				TagName: getTagName("v1.2.0-rc1"),
				Body:    enterpriseNotes(),
			},
		}
	})

	Context("Release Data", func() {
		var (
			releaseData *ReleaseData
		)
		JustBeforeEach(func() {
			var err error
			oReleaseData, err := NewMinorReleaseGroupedChangelogGenerator(Options{}, nil).NewReleaseData(openSourceReleases)
			Expect(err).NotTo(HaveOccurred())
			eReleaseData, err := NewMinorReleaseGroupedChangelogGenerator(Options{}, nil).NewReleaseData(enterpriseReleases)
			Expect(err).NotTo(HaveOccurred())
			releaseData, err = generator.MergeEnterpriseReleaseWithOS(eReleaseData, oReleaseData)
			Expect(err).NotTo(HaveOccurred())
		})

		It("groups minor versions correctly", func() {
			for _, release := range enterpriseReleases {
				releaseVersion, err := ParseVersion(release.GetTagName())
				Expect(err).NotTo(HaveOccurred())

				minorVersionGroup := GetMajorAndMinorVersion(releaseVersion)
				Expect(releaseData.Releases).To(HaveKey(minorVersionGroup))

				changelogNotes := releaseData.Releases[minorVersionGroup].ChangelogNotes
				Expect(changelogNotes).To(HaveKey(*releaseVersion))
			}
		})

		It("merges more than one open source versions correctly", func() {
			version := MustParseVersion("v1.2.1")
			noteCategories := releaseData.Releases[GetMajorAndMinorVersion(version)].ChangelogNotes[*version].Categories
			Expect(noteCategories).To(HaveKey("Helm Changes"))
			Expect(noteCategories["Helm Changes"]).To(ContainElement(&Note{Note: "Open Source Helm Change", FromDependentVersion: MustParseVersion("v1.2.0-rc2")}))
			Expect(noteCategories["Helm Changes"]).To(ContainElement(&Note{Note: "Open Source Helm Change", FromDependentVersion: MustParseVersion("v1.2.0")}))
		})

		It("includes no open source notes if there are no previous enterprise versions", func() {
			version := MustParseVersion("v1.2.0-beta13")
			noteCategories := releaseData.Releases[GetMajorAndMinorVersion(version)].ChangelogNotes[*version].Categories
			Expect(noteCategories).To(HaveKey("Helm Changes"))
			Expect(noteCategories["Helm Changes"]).To(ContainElement(&Note{Note: "Open Source Helm Change", FromDependentVersion: MustParseVersion("v1.1.0")}))
			Expect(noteCategories["Helm Changes"]).ToNot(ContainElement(&Note{Note: "Open Source Helm Change", FromDependentVersion: MustParseVersion("v1.1.0-beta9")}))
		})

		Context("Enterprise release data with no open source dependency", func() {
			BeforeEach(func() {
				enterpriseReleases = append(enterpriseReleases, &github.RepositoryRelease{
					TagName: getTagName("v1.2.0-rc6"),
					Body:    enterpriseNotes(),
				})
			})

			It("doesn't break, includes only enterprise notes because there is no open source dependency", func(){
				version := MustParseVersion("v1.2.0-rc6")
				noteCategories, _ := releaseData.Releases[GetMajorAndMinorVersion(version)].ChangelogNotes[*version].Dump()
				//Expect(noteCategories).To(HaveKey("Helm Changes"))
				print(noteCategories)
			})

		})
	})



})

func MustParseVersion(version string) *Version{
	v, err := ParseVersion(version)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return v
}
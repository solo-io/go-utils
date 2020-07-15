package versionutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/versionutils"
)

var _ = Describe("Version", func() {

	Context("matchesRegex", func() {
		It("works", func() {
			Expect(versionutils.MatchesRegex("v0.1.2")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v0.0.0")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v0.0.1")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v1.0.0")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v1.0.0-rc1")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v1.0.0-rc10")).To(BeTrue(), "should allow versions with 2 digits")
			Expect(versionutils.MatchesRegex("v0.5.20-rc100")).To(BeTrue(), "should allow verions with 3 digits")
			Expect(versionutils.MatchesRegex("v0.0.0-rc1")).To(BeTrue())
			Expect(versionutils.MatchesRegex("0.1.2")).To(BeFalse())
			Expect(versionutils.MatchesRegex("v1.2")).To(BeFalse())
			Expect(versionutils.MatchesRegex("v1.1.2.5")).To(BeFalse())
			Expect(versionutils.MatchesRegex("vX.Y.2")).To(BeFalse())
			Expect(versionutils.MatchesRegex("v1.2.3-rc")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v1.2.3-rc-1")).To(BeFalse())
			Expect(versionutils.MatchesRegex("v1.2.3-release1")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v1.2.3-beta1")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v1.2.3-")).To(BeFalse())
			Expect(versionutils.MatchesRegex("v1.2.3-1")).To(BeFalse())
			Expect(versionutils.MatchesRegex("v1.2.3+rc1")).To(BeFalse())
			Expect(versionutils.MatchesRegex("v1.2.3-beta")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v1.2.3-wasm")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v1.2.3-beta1-wasm")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v1.2.3-beta10-wasm")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v1.0.0-rc1-wasm")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v1.0.0-rc-wasm")).To(BeTrue())
			Expect(versionutils.MatchesRegex("vX.Y.2-wasm")).To(BeFalse())
			Expect(versionutils.MatchesRegex("v-wasm")).To(BeFalse())
			Expect(versionutils.MatchesRegex("-wasm")).To(BeFalse())
			Expect(versionutils.MatchesRegex("v1.2.3-foo")).To(BeTrue(), "should allow an unversioned label")
			Expect(versionutils.MatchesRegex("v1.2.3-foo-bar")).To(BeTrue(), "should allow hyphenated-labels")
			Expect(versionutils.MatchesRegex("v1.2.3-foo1-bar")).To(BeTrue(), "should allow one leading label version")
			Expect(versionutils.MatchesRegex("v1.2.3-foo-bar1")).To(BeTrue(), "should allow one trailing label version")
		})
	})

	Context("ParseVersion", func() {
		matches := func(tag string, major, minor, patch int, label string, labelVersion int) bool {
			parsed, err := versionutils.ParseVersion(tag)
			return err == nil && (*parsed == *versionutils.NewVersion(major, minor, patch, label, labelVersion))
		}

		It("works", func() {
			Expect(matches("v0.0.0", 0, 0, 0, "", 0)).To(BeTrue())
			Expect(matches("v0.1.2", 0, 1, 2, "", 0)).To(BeTrue())
			Expect(matches("v0.1.2", 0, 1, 3, "", 0)).To(BeFalse())
		})

		It("errors when invalid semver version provided", func() {
			parsed, err := versionutils.ParseVersion("0.1.2")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal(versionutils.InvalidSemverVersionError("0.1.2").Error()))
			Expect(parsed).To(BeNil())
		})

		It("supports wasm versions", func() {
			parsed, err := versionutils.ParseVersion("v1.5.0-rc1-wasm")
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.Label).To(Equal("rc-wasm"))
			Expect(parsed.LabelVersion).To(Equal(1))

			parsed, err = versionutils.ParseVersion("v1.5.0-beta2-wasm")
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.Label).To(Equal("beta-wasm"))
			Expect(parsed.LabelVersion).To(Equal(2))

			parsed, err = versionutils.ParseVersion("v1.5.1-wasm")
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.Label).To(Equal("wasm"))
			Expect(parsed.LabelVersion).To(Equal(0))
		})

		It("supports arbitrary trailing labels with a single version", func() {
			Expect(matches("v1.2.3-foo", 1, 2, 3, "foo", 0)).To(BeTrue())
			Expect(matches("v1.2.3-foo-bar", 1, 2, 3, "foo-bar", 0)).To(BeTrue())
			Expect(matches("v1.2.3-foo1-bar", 1, 2, 3, "foo-bar", 1)).To(BeTrue())
			Expect(matches("v1.2.3-foo-bar1", 1, 2, 3, "foo-bar", 1)).To(BeTrue())
		})

		It("rejects labels with multiple versions", func() {
			_, err := versionutils.ParseVersion("v1.2.3-foo1-bar1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("invalid label and version (multiple version numbers not allowed) foo1-bar1"))
		})

	})

	Context("IsGreaterThanTag", func() {

		expectResult := func(greater, lesser string, isGreaterThanOrEqualTo, determinable bool, err string) {
			actualWorked, determinable, actualErr := versionutils.IsGreaterThanTag(greater, lesser)
			Expect(actualWorked).To(BeEquivalentTo(isGreaterThanOrEqualTo))
			greaterVersion, parseGreaterErr := versionutils.ParseVersion(greater)
			lesserVersion, parseLesserErr := versionutils.ParseVersion(lesser)
			gteResult, determinable, gteError := greaterVersion.IsGreaterThanOrEqualToPtr(lesserVersion)
			if err == "" {
				Expect(actualErr).To(BeNil())
				Expect(gteResult).To(BeEquivalentTo(isGreaterThanOrEqualTo))
				Expect(gteError).To(BeNil())
			} else {
				Expect(actualErr.Error()).To(BeEquivalentTo(err))
			}
			if parseGreaterErr != nil {
				Expect(parseGreaterErr.Error()).To(BeEquivalentTo(err))
				Expect(gteError.Error()).To(BeEquivalentTo("cannot compare versions, greater version is nil"))
			}
			if parseLesserErr != nil {
				Expect(parseLesserErr.Error()).To(BeEquivalentTo(err))
				Expect(gteError.Error()).To(BeEquivalentTo("cannot compare versions, lesser version is nil"))
			}
		}

		It("works", func() {
			expectResult("v0.1.2", "v0.0.1", true, true, "")
			expectResult("v0.0.1", "v0.1.2", false, true, "")
			expectResult("v0.0.1", "v0.0.0", true, true, "")
			expectResult("0.0.2", "v0.0.1", false, true, versionutils.InvalidSemverVersionError("0.0.2").Error())
			expectResult("v0.0.2", "0.0.1", false, true, versionutils.InvalidSemverVersionError("0.0.1").Error())
			expectResult("v1.0.0", "v0.0.1-rc1", true, true, "")
			expectResult("v1.0.0-rc1", "v1.0.0-rc2", false, true, "")
			expectResult("v1.0.0-rc2", "v1.0.0-rc1", true, true, "")
			expectResult("v1.0.0-rc1", "v1.0.0", false, true, "")
			expectResult("v1.0.0", "v1.0.0-rc1", true, true, "")
			expectResult("v1.0.0-rc1", "v1.0.0-beta2", false, false, "")
			expectResult("v1.0.0-rc2", "v1.0.0-beta1", false, false, "")
			expectResult("v1.0.0-rc1", "v1.0.0-rc2", false, true, "")
			expectResult("v1.0.0-rc2", "v1.0.0-rc1", true, true, "")
			expectResult("v1.0.0-rc1-wasm", "v1.0.0-rc2-wasm", false, true, "")
			expectResult("v1.0.0-rc2-wasm", "v1.0.0-rc1-wasm", true, true, "")
			expectResult("v1.0.0-wasm", "v1.0.0", false, false, "")
			expectResult("v1.0.1-wasm", "v1.0.0-wasm", true, true, "")
			expectResult("v1.0.0-rc1-wasm", "v1.0.0-rc1", false, false, "")
			expectResult("v1.0.0-rc1", "v1.0.0-rc1-wasm", false, false, "")
			expectResult("v1.0.0-foo2-bar", "v1.0.1-foo1-bar", false, true, "")
			expectResult("v1.0.0-foo2-bar", "v1.0.0-foo1-bar", true, true, "")
		})
	})

	Describe("MustIsGreaterThan", func() {
		DescribeTable("it works", func(greater, lesser string, expected bool) {
			greaterVer, err := versionutils.ParseVersion(greater)
			Expect(err).NotTo(HaveOccurred())
			lesserVer, err := versionutils.ParseVersion(lesser)
			Expect(err).NotTo(HaveOccurred())
			Expect(greaterVer.MustIsGreaterThan(*lesserVer)).To(Equal(expected))
		},
			Entry("rc1 is greater than beta1", "v1.0.0-rc1", "v1.0.0-beta1", true),
			Entry("rc1 is greater than beta8", "v1.0.0-rc1", "v1.0.0-beta8", true),
			Entry("alpha is less than beta", "v1.0.0-alpha1", "v1.0.0-beta1", false),
			Entry("rc1 is not greater than itself", "v1.0.0-rc1", "v1.0.0-rc1", false),
		)
	})

	Context("GetVersionFromTag", func() {
		It("works", func() {
			Expect(versionutils.GetVersionFromTag("v0.1.2")).To(Equal("0.1.2"))
			Expect(versionutils.GetVersionFromTag("0.1.2")).To(Equal("0.1.2"))
		})
	})

	Context("IncrementVersion", func() {

		expectResult := func(start *versionutils.Version, breakingChange bool, newFeature bool, expected *versionutils.Version) {
			actualIncremented := start.IncrementVersion(breakingChange, newFeature)
			Expect(actualIncremented).To(BeEquivalentTo(expected))
		}

		getVersion := func(major, minor, patch int) *versionutils.Version {
			return versionutils.NewVersion(major, minor, patch, "", 0)
		}

		v0_0_1 := getVersion(0, 0, 1)
		v0_0_2 := getVersion(0, 0, 2)
		v0_1_0 := getVersion(0, 1, 0)
		v0_1_10 := getVersion(0, 1, 10)
		v0_1_11 := getVersion(0, 1, 11)
		v0_2_0 := getVersion(0, 2, 0)
		v1_1_10 := getVersion(1, 1, 10)
		v1_1_11 := getVersion(1, 1, 11)
		v1_2_0 := getVersion(1, 2, 0)
		v2_0_0_foo_1 := versionutils.NewVersion(2, 0, 0, "foo", 1)
		v2_0_0_foo_2 := versionutils.NewVersion(2, 0, 0, "foo", 2)
		v2_0_0 := getVersion(2, 0, 0)
		v2_0_0_wasm_1 := versionutils.NewVersion(2, 0, 0, "wasm", 1)
		v2_0_0_wasm_2 := versionutils.NewVersion(2, 0, 0, "wasm", 2)
		v2_0_0_foo_wasm_1 := versionutils.NewVersion(2, 0, 0, "foo-wasm", 1)
		v2_0_0_foo_wasm_2 := versionutils.NewVersion(2, 0, 0, "foo-wasm", 2)

		It("works", func() {
			expectResult(v0_0_1, true, true, v0_1_0)
			expectResult(v0_0_1, true, false, v0_1_0)
			expectResult(v0_1_10, true, false, v0_2_0)
			expectResult(v0_1_10, true, true, v0_2_0)
			expectResult(v1_1_10, true, false, v2_0_0)
			expectResult(v1_1_10, true, true, v2_0_0)
			expectResult(v0_0_1, false, false, v0_0_2)
			expectResult(v0_0_1, false, true, v0_0_2)
			expectResult(v0_1_10, false, false, v0_1_11)
			expectResult(v0_1_10, false, true, v0_1_11)
			expectResult(v1_1_10, false, false, v1_1_11)
			expectResult(v1_1_10, false, true, v1_2_0)
			expectResult(v2_0_0_foo_1, false, true, v2_0_0_foo_2)
			expectResult(v2_0_0_wasm_1, false, true, v2_0_0_wasm_2)
			expectResult(v2_0_0_foo_wasm_1, false, true, v2_0_0_foo_wasm_2)
		})
	})

})

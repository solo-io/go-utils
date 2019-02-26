package versionutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/versionutils"
)

var _ = Describe("Version", func() {

	getVersion := func(major, minor, patch int) *versionutils.Version {
		return &versionutils.Version{
			Major: major,
			Minor: minor,
			Patch: patch,
		}
	}

	var _ = Context("matchesRegex", func() {
		It("works", func() {
			Expect(versionutils.MatchesRegex("v0.1.2")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v0.0.0")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v0.0.1")).To(BeTrue())
			Expect(versionutils.MatchesRegex("v1.0.0")).To(BeTrue())
			Expect(versionutils.MatchesRegex("0.1.2")).To(BeFalse())
			Expect(versionutils.MatchesRegex("v1.2")).To(BeFalse())
			Expect(versionutils.MatchesRegex("v1.1.2.5")).To(BeFalse())
			Expect(versionutils.MatchesRegex("vX.Y.2")).To(BeFalse())
		})
	})

	var _ = Context("ParseVersion", func() {
		matches := func(tag string, major, minor, patch int) bool {
			parsed, err := versionutils.ParseVersion(tag)
			return err == nil && (*parsed == *getVersion(major, minor, patch))
		}

		It("works", func() {
			Expect(matches("v0.1.2", 0, 1, 2)).To(BeTrue())
			Expect(matches("v0.1.2", 0, 1, 3)).To(BeFalse())
		})

		It("errors when invalid semver version provided", func() {
			parsed, err := versionutils.ParseVersion("0.1.2")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(BeEquivalentTo("Tag 0.1.2 is not a valid semver version, must be of the form vX.Y.Z"))
			Expect(parsed).To(BeNil())
		})

		It("errors when zero version provided", func() {
			parsed, err := versionutils.ParseVersion("v0.0.0")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(BeEquivalentTo("Version v0.0.0 is not greater than v0.0.0"))
			Expect(parsed).To(BeNil())
		})

	})

	var _ = Context("IsGreaterThanTag", func() {

		expectResult := func(greater, lesser string, worked bool, err string) {
			actualWorked, actualErr := versionutils.IsGreaterThanTag(greater, lesser)
			Expect(actualWorked).To(BeEquivalentTo(worked))
			if err == "" {
				Expect(actualErr).To(BeNil())
			} else {
				Expect(actualErr.Error()).To(BeEquivalentTo(err))
			}
		}

		It("works", func() {
			expectResult("v0.1.2", "v0.0.1", true, "")
			expectResult("v0.0.1", "v0.1.2", false, "")
			expectResult("v0.0.1", "v0.0.0", false, "Version v0.0.0 is not greater than v0.0.0")
			expectResult("0.0.2", "v0.0.1", false, "Tag 0.0.2 is not a valid semver version, must be of the form vX.Y.Z")
			expectResult("v0.0.2", "0.0.1", false, "Tag 0.0.1 is not a valid semver version, must be of the form vX.Y.Z")
		})
	})

})

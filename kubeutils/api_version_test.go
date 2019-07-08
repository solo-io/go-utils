package kubeutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
)

var _ = Describe("ApiVersion", func() {
	Describe("ParseApiVersion", func() {
		DescribeTable("it works", func(in string, expected kubeutils.ApiVersion) {
			actual, err := kubeutils.ParseApiVersion(in)
			Expect(err).NotTo(HaveOccurred())
			Expect(actual.Equal(expected)).To(BeTrue())
		},
			Entry("valid minimal version", "v1", kubeutils.NewApiVersion(1, 0, kubeutils.GA)),
			Entry("valid 2-digit version", "v11", kubeutils.NewApiVersion(11, 0, kubeutils.GA)),
			Entry("valid alpha version", "v1alpha1", kubeutils.NewApiVersion(1, 1, kubeutils.Alpha)),
			Entry("valid beta version", "v11beta11", kubeutils.NewApiVersion(11, 11, kubeutils.Beta)),
		)

		DescribeTable("it errors", func(in string, expectedErr error) {
			actual, err := kubeutils.ParseApiVersion(in)
			Expect(actual).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedErr.Error()))
		},
			Entry("no v", "1", kubeutils.MalformedVersionError("1")),
			Entry("zero version", "v0", kubeutils.InvalidMajorVersionError),
			Entry("zero alpha", "v1alpha0", kubeutils.InvalidPrereleaseVersionError),
		)
	})

	Describe("accessors", func() {
		var subject kubeutils.ApiVersion
		BeforeEach(func() {
			subject = kubeutils.NewApiVersion(1, 2, kubeutils.Beta)
		})

		Describe("Major", func() {
			It("works", func() {
				Expect(subject.Major()).To(Equal(1))
			})
		})

		Describe("Prerelease", func() {
			It("works", func() {
				Expect(subject.Prerelease()).To(Equal(2))
			})
		})

		Describe("PrereleaseModifier", func() {
			It("works", func() {
				Expect(subject.PrereleaseModifier()).To(Equal(kubeutils.Beta))
			})
		})
	})

	Describe("String", func() {
		DescribeTable("it works", func(str string) {
			subject, err := kubeutils.ParseApiVersion(str)
			Expect(err).NotTo(HaveOccurred())
			Expect(subject.String()).To(Equal(str))
		},
			Entry("minimal version", "v1"),
			Entry("2-digit version", "v11"),
			Entry("alpha version", "v1alpha1"),
			Entry("beta version", "v11beta11"),
		)
	})

})

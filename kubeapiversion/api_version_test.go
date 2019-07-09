package kubeapiversion

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("ApiVersion", func() {
	Describe("ParseApiVersion", func() {
		DescribeTable("it works", func(in string, expected ApiVersion) {
			actual, err := ParseApiVersion(in)
			Expect(err).NotTo(HaveOccurred())
			Expect(actual.Equal(expected)).To(BeTrue())
		},
			Entry("valid minimal version", "v1", &apiVersion{major: 1, prerelease: 0, modifier: GA}),
			Entry("valid 2-digit version", "v11", &apiVersion{major: 11, prerelease: 0, modifier: GA}),
			Entry("valid alpha version", "v1alpha1", &apiVersion{major: 1, prerelease: 1, modifier: Alpha}),
			Entry("valid beta version", "v11beta11", &apiVersion{major: 11, prerelease: 11, modifier: Beta}),
		)

		DescribeTable("it errors", func(in string, expectedErr error) {
			actual, err := ParseApiVersion(in)
			Expect(actual).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedErr.Error()))
		},
			Entry("no v prefix", "1", MalformedVersionError("1")),
			Entry("non-v prefix", "x1", MalformedVersionError("x1")),
			Entry("invalid modifier", "1gamma1", MalformedVersionError("1gamma1")),
			Entry("trailing non-numerals", "1alpha2abc", MalformedVersionError("1alpha2abc")),
			Entry("non-alphanumeric", "1alpha!2", MalformedVersionError("1alpha!2")),
			Entry("not even dashes", "1alpha-2", MalformedVersionError("1alpha-2")),
			Entry("zero version", "v0", InvalidMajorVersionError),
			Entry("zero alpha", "v1alpha0", InvalidPrereleaseVersionError),
			Entry("zero beta", "v1beta0", InvalidPrereleaseVersionError),
		)
	})

	Describe("accessors", func() {
		var subject ApiVersion
		BeforeEach(func() {
			subject = &apiVersion{major: 1, prerelease: 2, modifier: Beta}
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
				Expect(subject.PrereleaseModifier()).To(Equal(Beta))
			})
		})
	})

	Describe("String", func() {
		DescribeTable("it works", func(str string) {
			subject, err := ParseApiVersion(str)
			Expect(err).NotTo(HaveOccurred())
			Expect(subject.String()).To(Equal(str))
		},
			Entry("minimal version", "v1"),
			Entry("2-digit version", "v11"),
			Entry("alpha version", "v1alpha1"),
			Entry("beta version", "v11beta11"),
		)
	})

	DescribeTable("GreaterThan", func(a, b string, expected bool) {
		subject, err := ParseApiVersion(a)
		Expect(err).NotTo(HaveOccurred())
		other, err := ParseApiVersion(b)
		Expect(err).NotTo(HaveOccurred())
		Expect(subject.GreaterThan(other)).To(Equal(expected))
	},
		Entry("greater major", "v2", "v1", true),
		Entry("greater major vs alpha", "v2", "v1alpha2", true),
		Entry("greater major vs beta", "v2", "v1beta2", true),
		Entry("greater alpha", "v1alpha2", "v1alpha1", true),
		Entry("greater beta", "v1beta2", "v1beta1", true),
		Entry("beta vs alpha", "v1beta1", "v1alpha1", true),
		Entry("equal", "v1beta1", "v1beta1", false),
		Entry("lesser major", "v1", "v2", false),
		Entry("lesser major with alpha", "v1alpha1", "v2", false),
		Entry("lesser major with beta", "v1beta1", "v2", false),
		Entry("lesser alpha", "v1alpha1", "v1alpha2", false),
		Entry("lesser beta", "v1beta1", "v1beta2", false),
		Entry("alpha vs beta", "v1alpha1", "v1beta1", false),
	)

	DescribeTable("LessThan", func(a, b string, expected bool) {
		subject, err := ParseApiVersion(a)
		Expect(err).NotTo(HaveOccurred())
		other, err := ParseApiVersion(b)
		Expect(err).NotTo(HaveOccurred())
		Expect(subject.LessThan(other)).To(Equal(expected))
	},
		Entry("greater major", "v2", "v1", false),
		Entry("greater major vs alpha", "v2", "v1alpha2", false),
		Entry("greater major vs beta", "v2", "v1beta2", false),
		Entry("greater alpha", "v1alpha2", "v1alpha1", false),
		Entry("greater beta", "v1beta2", "v1beta1", false),
		Entry("beta vs alpha", "v1beta1", "v1alpha1", false),
		Entry("equal", "v1beta1", "v1beta1", false),
		Entry("lesser major", "v1", "v2", true),
		Entry("lesser major with alpha", "v1alpha1", "v2", true),
		Entry("lesser major with beta", "v1beta1", "v2", true),
		Entry("lesser alpha", "v1alpha1", "v1alpha2", true),
		Entry("lesser beta", "v1beta1", "v1beta2", true),
		Entry("alpha vs beta", "v1alpha1", "v1beta1", true),
	)

	DescribeTable("Equal", func(a, b string, expected bool) {
		subject, err := ParseApiVersion(a)
		Expect(err).NotTo(HaveOccurred())
		other, err := ParseApiVersion(b)
		Expect(err).NotTo(HaveOccurred())
		Expect(subject.Equal(other)).To(Equal(expected))
	},
		Entry("major", "v1", "v1", true),
		Entry("alpha", "v1alpha1", "v1alpha1", true),
		Entry("beta", "v1beta1", "v1beta1", true),
		Entry("many digits", "v111beta222", "v111beta222", true),
		Entry("major mismatch", "v1", "v2", false),
		Entry("major mismatch with alpha", "v1alpha1", "v2alpha1", false),
		Entry("alpha mismatch", "v1alpha2", "v1alpha1", false),
		Entry("beta mismatch", "v1beta2", "v1beta1", false),
		Entry("prerelease mismatch", "v1alpha1", "v1beta1", false),
	)
})

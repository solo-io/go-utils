package kubeutils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("sanitize name", func() {

	DescribeTable("sanitize short names", func(in, out string) {
		Expect(SanitizeNameV2(in)).To(Equal(out))
	},
		Entry("basic a", "abc", "abc"),
		Entry("basic b", "abc123", "abc123"),
		Entry("subX *", "bb*", "bb-"),
		Entry("sub *", "bb*b", "bb-b"),
		Entry("subX /", "bb/", "bb-"),
		Entry("sub /", "bb/b", "bb-b"),
		Entry("subX .", "bb.", "bb-"),
		Entry("sub .", "bb.b", "bb-b"),
		Entry("sub0 [", "bb[", "bb"),
		Entry("sub [", "bb[b", "bbb"),
		Entry("sub0 ]", "bb]", "bb"),
		Entry("sub ]", "bb]b", "bbb"),
		Entry("subX :", "bb:", "bb-"),
		Entry("sub :", "bb:b", "bb-b"),
		Entry("subX space", "bb ", "bb-"),
		Entry("sub space", "bb b", "bb-b"),
		Entry("subX newline", "bb\n", "bb"),
		Entry("sub newline", "bb\nb", "bbb"),
		Entry("sub0 quote", "aa\"", "aa"),
		Entry("sub quote b", "bb\"b", "bbb"),
		Entry("sub0 single quote", "aa'", "aa"),
		Entry("sub single quote b", "bb'b", "bbb"),
		// these are technically invalid kube names, as are the subX cases, but user should know that and kube wil warn
		Entry("invalid a", "123", "123"),
		Entry("invalid b", "-abc", "-abc"),
	)

	DescribeTable("sanitize long names", func(in, out string) {
		sanitized := SanitizeNameV2(in)
		Expect(sanitized).To(Equal(out))
		Expect(len(sanitized)).To(BeNumerically("<=", 63))
	},
		Entry("300a's", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-4e5475d125a33c6190718e75adc1b70"),
		Entry("301a's", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-c73301b7b71679067b02cff4cdc5e70"),
	)
})

package common_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/v1/common"
)

var _ = Describe("RandStringBytes", func() {
	It("Handle zero-length input in a reasonable way", func() {
		Expect(common.RandStringBytes(0, "")).To(Equal(""))
		Expect(common.RandStringBytes(0, "a")).To(Equal(""))
	})

	It("Should generate random strings: length = 1", func() {
		Expect(common.RandStringBytes(1, "")).To(Equal(""))
		Expect(common.RandStringBytes(1, "a")).To(Equal("a"))
	})

	It("Should generate random strings: length > 1", func() {
		Expect(common.RandStringBytes(2, "aaa")).To(Equal("aa"))
		Expect(common.RandStringBytes(1000, "ab")).To(MatchRegexp("a"))
		Expect(common.RandStringBytes(2, "ab")).ToNot(MatchRegexp("c"))
	})
})

var _ = Describe("RandKubeNameBytes", func() {
	It("Handle invalid input in a reasonable way", func() {
		Expect(common.RandKubeNameBytes(0)).To(Equal(""))
	})

	It("Should generate valid Kube names", func() {
		for i := 1; i < 20; i++ {
			Expect(common.RandKubeNameBytes(i)).To(MatchRegexp("[a-z]([-a-z0-9]*[a-z0-9])?"))
		}
	})
})

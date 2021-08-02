package stringutils_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/stringutils"
)

var _ = Describe("stringutils", func() {
	Context("MapStringSlice", func() {
		var (
			stringSlice []string
		)

		BeforeEach(func() {
			stringSlice = []string{"a", "b", "c", "d", "e"}
		})

		It("works", func() {
			transformedSlice := stringutils.MapStringSlice(stringSlice, strings.ToUpper)
			Expect(transformedSlice).To(Equal([]string{"A", "B", "C", "D", "E"}))
			Expect(stringSlice).To(Equal([]string{"a", "b", "c", "d", "e"}))
		})
	})

	Context("StringSliceToHashSet", func() {
		var (
			stringSlice []string
		)

		BeforeEach(func() {
			stringSlice = []string{"a", "b", "c", "d", "e", "e"}
		})

		It("Works", func() {
			hashSet := stringutils.StringSliceToHashSet(stringSlice)
			for _, s := range stringSlice {
				Expect(hashSet).To(HaveKeyWithValue(s, true))
			}
			Expect(stringSlice).To(Equal([]string{"a", "b", "c", "d", "e", "e"}))
		})
	})
})

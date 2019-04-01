package surveyutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/surveyutils"
	clitestutils "github.com/solo-io/go-utils/testutils/cli"
)

var _ = Describe("GetInput", func() {
	Context("bool input", func() {
		It("correctly sets the input value", func() {
			clitestutils.ExpectInteractive(func(c *clitestutils.Console) {
				c.ExpectString("test msg [y/N]: ")
				c.SendLine("y")
				c.ExpectEOF()
			}, func() {
				var val bool
				err := surveyutils.GetBoolInput("test msg", &val)
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(BeTrue())
			})
		})
	})

	Context("list select", func() {
		var options = []string{"one", "two", "three"}
		Context("single select", func() {
			It("can select 1 from list", func() {
				clitestutils.ExpectInteractive(func(c *clitestutils.Console) {
					c.ExpectString("select option")
					c.PressDown()
					c.SendLine("")
					c.ExpectEOF()
				}, func() {
					var val string
					err := surveyutils.ChooseFromList("select option", &val, options)
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(Equal("two"))
				})
			})
		})
		Context("multi select", func() {
			It("can select one from a mutli-select list", func() {
				clitestutils.ExpectInteractive(func(c *clitestutils.Console) {
					c.ExpectString("select option")
					c.PressDown()
					c.Send(" ")
					c.SendLine("")
					c.ExpectEOF()
				}, func() {
					var val []string
					err := surveyutils.ChooseMultiFromList("select option", &val, options)
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(Equal([]string{"two"}))
				})
			})

			It("can select mutli from a mutli-select list", func() {
				clitestutils.ExpectInteractive(func(c *clitestutils.Console) {
					c.ExpectString("select option")
					c.PressDown()
					c.Send(" ")
					c.PressDown()
					c.Send(" ")
					c.SendLine("")
					c.ExpectEOF()
				}, func() {
					var val []string
					err := surveyutils.ChooseMultiFromList("select option", &val, options)
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(Equal([]string{"two", "three"}))
				})
			})
		})
	})
})

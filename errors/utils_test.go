package errors

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("new golang errors", func() {
	var (
		baseError     = New("base error")
		wrapperError1 = Wrapf(baseError, "wrapper 1")
		wrapperError2 = Wrapf(wrapperError1, "wrapper 2")
	)

	It("can compare underlyin error values", func() {
		Expect(Is(wrapperError1, baseError)).To(BeTrue())
		Expect(Is(wrapperError2, baseError)).To(BeTrue())
		Expect(Is(wrapperError2, wrapperError1)).To(BeTrue())
	})
})

package testutils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
)

var _ = Describe("eris errors", func() {
	var (
		baseError     = eris.New("base error")
		wrapperError1 = eris.Wrap(baseError, "wrapper 1")
		wrapperError2 = eris.Wrap(wrapperError1, "wrapper 2")
	)

	It("can compare underlying error values", func() {
		Expect(baseError).To(HaveInErrorChain(baseError), "Identity should work")
		Expect(wrapperError1).To(HaveInErrorChain(baseError), "Chaining should work")
		Expect(wrapperError2).To(HaveInErrorChain(baseError), "Chaining should work at any depth")
		Expect(wrapperError2).To(HaveInErrorChain(wrapperError1), "Chaining should pick up intermediate errors")
		Expect(wrapperError2).NotTo(HaveInErrorChain(eris.New("dummy error")), "Should not pick up errors that are not in the chain")
		Expect(wrapperError2).NotTo(HaveInErrorChain(nil), "nil as the expected error should not blow up")
		Expect(nil).NotTo(HaveInErrorChain(baseError), "nil as the actual error should not blow up")
	})
})

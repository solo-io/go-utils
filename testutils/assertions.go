package testutils

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

// ExpectEqualProtoMessages provides richer error messages than struct comparison by leveraging the String() method that all
// proto Messages provide. On error, Gomega's string comparison utility prints a few characters of the text immediately
// surrounding the first discrepancy. Customize the size of the rendered diff by setting format.CharactersAroundMismatchToInclude
//
// Variadic optionalDescription argument is passed on to fmt.Sprintf() and is used to annotate failure messages.
//
// Example of the output:
//   optionalDescription is rendered here: template string with values foo and bar
//   Expected
//       <string>: "...-1010" vers..."
//   to equal               |
//       <string>: "...-10101" ver..."
func ExpectEqualProtoMessages(a, b proto.Message, optionalDescription ...interface{}) {
	if proto.Equal(a, b) {
		fmt.Fprintf(ginkgo.GinkgoWriter, "protos do not match\nhave (length %v):\n%v\n want (length %v):\n %v", len(a.String()), a.String(), len(b.String()), b.String())
		fmt.Fprintf(ginkgo.GinkgoWriter, "set CharactersAroundMismatchToInclude to include more or less diff context (current value is %v)", format.CharactersAroundMismatchToInclude)
	}

	Expect(a.String()).To(Equal(b.String()), optionalDescription...)
}

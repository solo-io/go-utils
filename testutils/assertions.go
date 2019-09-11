package testutils

import (
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

// ExpectEqualProtoMessages provides richer error messages than struct comparison by leveraging the String() method that all
// proto Messages provide. On error, Gomega's string comparison utility prints the characters of the text immediately
// surrounding the first discrepancy. The number of characters rendered is determined by format.CharactersAroundMismatchToInclude.
// Here, we set the value to 200 to highlight the diff in context without overwhelming users with too much information.
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
		return
	}

	initialCharactersAroundMismatchToInclude := format.CharactersAroundMismatchToInclude
	format.CharactersAroundMismatchToInclude = 200
	defer func() {
		format.CharactersAroundMismatchToInclude = initialCharactersAroundMismatchToInclude
	}()

	Expect(a.String()).To(Equal(b.String()), optionalDescription...)
}

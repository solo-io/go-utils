package testutils

import (
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/gomega"
)

// ExpectEqualProtoMessages provides richer error messages than struct comparison by leveraging the String() method that all
// proto Messages provide. On error, Gomega's string comparison utility prints a few characters of the text immediately
// surrounding the first discrepancy.
// Example of the output:
//   Expected
//       <string>: "...-1010" vers..."
//   to equal               |
//       <string>: "...-10101" ver..."
func ExpectEqualProtoMessages(a, b proto.Message) {
	if proto.Equal(a, b) {
		return
	}
	// One shortcoming is that you only get +/- 5 chars of context
	// per: https://github.com/onsi/gomega/blob/master/format/format.go#L146
	// TODO(mitchdraft) gomega pr to modify charactersAroundMismatchToInclude (if not merged will make a util)
	Expect(a.String()).To(Equal(b.String()))
}

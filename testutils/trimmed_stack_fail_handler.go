package testutils

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"runtime/debug"
)

// PrintTrimmedStack helps you find the line of the failing assertion without producing excessive noise.
// This is achieved by printing a stack trace and pruning lines associated with known overhead.
// With this fail handler, you do not need to count stack offsets ExpectWithOffset(x, ...) and can just Expect(...)
func PrintTrimmedStack() {
	stack := debug.Stack()
	fmt.Println(trimVendorStack(stack))
}
func trimVendorStack(stack []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(stack))
	ind := -1
	pair := []string{}
	skipCount := 0
	output := ""
	for scanner.Scan() {
		ind++
		if ind == 0 {
			// skip the header
			continue
		}
		pair = append(pair, scanner.Text())
		if len(pair) == 2 {
			evaluateStackPair(pair[0], pair[1], &output, &skipCount)
			pair = []string{}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	output = fmt.Sprintf("Stack trace (skipped %v entries that matched filter criteria):\n%v", skipCount, output)
	return output
}

var (
	funcRuntimeDebugRegex = &regexp.Regexp{}
	fileVendorRegex       = &regexp.Regexp{}
	fileGoModRegex        = &regexp.Regexp{}
	fileSuiteRegex        = &regexp.Regexp{}
	fileGoTestLibRegex    = &regexp.Regexp{}
	fileSelfDescription   = &regexp.Regexp{}
)

func init() {
	funcRuntimeDebugRegex = regexp.MustCompile("runtime/debug")
	fileVendorRegex = regexp.MustCompile("vendor")
	fileGoModRegex = regexp.MustCompile("/go/pkg/mod/")
	fileSuiteRegex = regexp.MustCompile("suite_test.go")
	fileGoTestLibRegex = regexp.MustCompile("src/testing/testing.go")
	fileSelfDescription = regexp.MustCompile("solo-io/go-utils/testutils/trimmed")
}

func evaluateStackPair(functionLine, fileLine string, output *string, skipCount *int) {
	skip := false
	if funcRuntimeDebugRegex.MatchString(functionLine) {
		skip = true
	}
	if fileVendorRegex.MatchString(fileLine) ||
		fileGoModRegex.MatchString(fileLine) ||
		fileSuiteRegex.MatchString(fileLine) ||
		fileGoTestLibRegex.MatchString(fileLine) ||
		fileSelfDescription.MatchString(fileLine) {
		skip = true
	}
	if skip {
		*skipCount = *skipCount + 1
		return
	}
	*output = fmt.Sprintf("%v%v\n%v\n", *output, functionLine, fileLine)
	return
}

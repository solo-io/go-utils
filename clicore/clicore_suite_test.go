package clicore

import (
	"bufio"
	"bytes"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/solo-io/solo-kit/test/helpers"
	"os"
	"regexp"
	"runtime/debug"
	"testing"
)

func TestInstall(t *testing.T) {

	helpers.RegisterPreFailHandler(
		func() {
			printTrimmedStack()
		})
	helpers.RegisterCommonFailHandlers()
	helpers.SetupLog()
	RunSpecs(t, "Clicore Suite")
}

// TODO(mitchdraft) - move to go-utils https://github.com/solo-io/go-utils/issues/131

func printTrimmedStack() {
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
	fileSuiteRegex        = &regexp.Regexp{}
	fileGoTestLibRegex    = &regexp.Regexp{}
)

func init() {
	funcRuntimeDebugRegex = regexp.MustCompile("runtime/debug")
	fileVendorRegex = regexp.MustCompile("vendor")
	fileSuiteRegex = regexp.MustCompile("suite_test.go")
	fileGoTestLibRegex = regexp.MustCompile("src/testing/testing.go")
}

func evaluateStackPair(functionLine, fileLine string, output *string, skipCount *int) {
	skip := false
	if funcRuntimeDebugRegex.MatchString(functionLine) {
		skip = true
	}
	if fileVendorRegex.MatchString(fileLine) ||
		fileSuiteRegex.MatchString(fileLine) ||
		fileGoTestLibRegex.MatchString(fileLine) {
		skip = true
	}
	if skip {
		*skipCount = *skipCount + 1
		return
	}
	*output = fmt.Sprintf("%v%v\n%v\n", *output, functionLine, fileLine)
	return
}

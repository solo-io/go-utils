package errors

import (
	"errors"
	"fmt"

	gomegatypes "github.com/onsi/gomega/types"
	"github.com/rotisserie/eris"
)

// Relies on eris (https://github.com/rotisserie/eris)'s concept of error identity, specifically:
//
// "eris.Is returns true if a particular error appears anywhere in the error chain...
// eris.Is works simply by comparing error messages with each other. If an error contains a
// particular message anywhere in its chain (e.g. "not found"), it's defined to be that error
// type (i.e. eris.Is will return true)."
//
// Example usage:
// Expect(wrapperError1).To(HaveInErrorChain(baseError), "Chaining should work")
func HaveInErrorChain(err error) gomegatypes.GomegaMatcher {
	return &errorChainMatcher{expected: err}
}

type errorChainMatcher struct {
	expected error
}

func (e *errorChainMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil {
		return false, nil
	}

	actualError, ok := actual.(error)
	if !ok {
		return false, errors.New("could not convert actual value to error type")
	}
	return eris.Is(actualError, e.expected), nil
}

func (e *errorChainMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Received actual: %+v, but expected error %+v", actual, e.expected)
}

func (e *errorChainMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected actual: %+v not to equal error +%v", actual, e.expected)
}

package testutils

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var preFails []func()

func RegisterPreFailHandler(prefail func()) {
	preFails = append(preFails, prefail)
}

func RegisterCommonFailHandlers() {
	RegisterPreFailHandler(waitOnFail)
	RegisterFailHandler(failHandler)
}

func failHandler(message string, callerSkip ...int) {
	fmt.Println("Fail handler msg", message)

	for _, preFail := range preFails {
		preFail()
	}

	// Account for this extra function in the call stack.
	// Without this all failure messages will show the incorrect line number!
	var shiftedCallerSkip []int
	for _, i := range callerSkip {
		shiftedCallerSkip = append(shiftedCallerSkip, i+1)
	}

	Fail(message, shiftedCallerSkip...)
}

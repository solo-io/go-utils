package testutils

import (
	"fmt"

	. "github.com/onsi/ginkgo"
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

	for _, prefail := range preFails {
		prefail()
	}
	Fail(message, callerSkip...)

}

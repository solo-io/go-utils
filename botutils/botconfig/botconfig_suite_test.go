package botconfig_test

import (
	"testing"

	"github.com/solo-io/go-utils/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBotconfig(t *testing.T) {
	test = t
	RegisterFailHandler(Fail)
	testutils.RegisterPreFailHandler(
		func() {
			testutils.PrintTrimmedStack()
		})
	testutils.RegisterCommonFailHandlers()
	RunSpecs(t, "Botconfig Suite")
}

var test *testing.T

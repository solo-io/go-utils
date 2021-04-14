package botconfig_test

import (
	"testing"

	"github.com/onsi/ginkgo/reporters"

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
	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Botconfig Suite", []Reporter{junitReporter})
}

var test *testing.T

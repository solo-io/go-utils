package changelogutils_test

import (
	"github.com/onsi/ginkgo/reporters"
	"testing"

	"github.com/solo-io/go-utils/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestChangelogUtils(t *testing.T) {
	test = t
	RegisterFailHandler(Fail)
	testutils.RegisterPreFailHandler(
		func() {
			testutils.PrintTrimmedStack()
		})
	testutils.RegisterCommonFailHandlers()
	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "ChangelogUtils Suite", []Reporter{junitReporter})
}

var test *testing.T

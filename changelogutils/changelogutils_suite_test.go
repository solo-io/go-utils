package changelogutils_test

import (
	"testing"

	"github.com/solo-io/go-utils/testutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestChangelogUtils(t *testing.T) {
	test = t
	RegisterFailHandler(Fail)
	testutils.RegisterPreFailHandler(testutils.PrintTrimmedStack)
	testutils.RegisterCommonFailHandlers()
	RunSpecs(t, "ChangelogUtils Suite")
}

var test *testing.T

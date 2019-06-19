package changelogutils_test

import (
	"testing"

	"github.com/solo-io/go-utils/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestChangelogUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	testutils.RegisterPreFailHandler(
		func() {
			testutils.PrintTrimmedStack()
		})
	testutils.RegisterCommonFailHandlers()
	RunSpecs(t, "ChangelogUtils Suite")
}

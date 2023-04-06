package versionutils_test

import (
	"testing"

	"github.com/solo-io/go-utils/testutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVersionUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	testutils.RegisterPreFailHandler(testutils.PrintTrimmedStack)
	testutils.RegisterCommonFailHandlers()
	RunSpecs(t, "Versionutils Suite")
}

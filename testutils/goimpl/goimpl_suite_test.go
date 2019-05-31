package goimpl

import (
	"testing"

	"github.com/solo-io/go-utils/testutils"

	. "github.com/onsi/ginkgo"
)

func TestGoImpl(t *testing.T) {

	testutils.RegisterPreFailHandler(
		func() {
			testutils.PrintTrimmedStack()
		})
	testutils.RegisterCommonFailHandlers()
	testutils.SetupLog()
	RunSpecs(t, "Go Impl Suite")
}

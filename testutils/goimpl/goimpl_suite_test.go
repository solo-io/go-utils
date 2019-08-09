package goimpl

import (
	"testing"

	"github.com/solo-io/go-utils/testutils"

	. "github.com/onsi/ginkgo"
)

func TestGoImpl(t *testing.T) {

	testutils.RegisterPreFailHandler(testutils.PrintTrimmedStack)
	testutils.RegisterCommonFailHandlers()
	RunSpecs(t, "Go Impl Suite")
}

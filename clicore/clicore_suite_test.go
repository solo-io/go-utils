package clicore

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
)

func TestCliCore(t *testing.T) {
	testutils.RegisterPreFailHandler(testutils.PrintTrimmedStack)
	testutils.RegisterCommonFailHandlers()
	RegisterFailHandler(Fail)
	testutils.SetupLog()
	RunSpecs(t, "Clicore Suite")
}

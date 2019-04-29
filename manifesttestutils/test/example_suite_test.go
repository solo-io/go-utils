package test

import (
	"testing"

	. "github.com/solo-io/go-utils/manifesttestutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestManifestTestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ManifestTestUtils Suite")
}

var (
	testManifest TestManifest
)

var _ = BeforeSuite(func() {
	testManifest = NewTestManifest("example.yaml")
})

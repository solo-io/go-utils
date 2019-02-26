package versionutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVersionUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Versionutils Suite")
}

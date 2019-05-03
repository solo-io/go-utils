package vfsutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVfsutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vfsutils Suite")
}

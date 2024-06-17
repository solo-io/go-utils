package testutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTestutilsBlackbox(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Testutils Suite")
}

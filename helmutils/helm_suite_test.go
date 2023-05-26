package helmutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHealthchecker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helm Suite")
}

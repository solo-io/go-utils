package loggingutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLoggingutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Loggingutils Suite")
}

package tarutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTarutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tarutils Suite")
}

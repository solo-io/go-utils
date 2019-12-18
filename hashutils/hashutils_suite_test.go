package hashutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHashutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hashutils Suite")
}

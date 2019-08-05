package builders

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	T *testing.T
)

func TestBuilders(t *testing.T) {
	T = t
	RegisterFailHandler(Fail)
	RunSpecs(t, "Builders Suite")
}

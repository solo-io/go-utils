package helminstall_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHelminstall(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helminstall Suite")
}

package pkgmgmtutils

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPkgmgmtutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pkgmgmtutils Suite")
}

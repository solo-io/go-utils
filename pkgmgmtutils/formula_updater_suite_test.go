package pkgmgmtutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFormulaUpdater(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FormulaUpdater Suite")
}

package brew_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFormulaUpdater(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FormulaUpdater Suite")
}

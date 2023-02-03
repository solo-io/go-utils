package stringutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSurveyutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stringutils Suite")
}

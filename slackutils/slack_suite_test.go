package slackutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSlackutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Slackutils Suite")
}

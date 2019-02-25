package changelogutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestChangelogUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ChangelogUtils Suite")
}

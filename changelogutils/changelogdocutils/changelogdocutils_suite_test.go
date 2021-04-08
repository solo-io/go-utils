package changelogdocutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestChangelogDocUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ChangelogDocUtils Suite")
}

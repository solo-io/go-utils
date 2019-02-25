package githubutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGithubutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Githubutils Suite")
}

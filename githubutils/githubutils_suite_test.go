package githubutils_test

import (
    "testing"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

func TestGithubutils(t *testing.T) {
    // TODO (sam-h)
    Skip("Temporarily skip tests as they are known to be rate-limited")
    RegisterFailHandler(Fail)
    RunSpecs(t, "Githubutils Suite")
}

package gcloudutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCloudbuild(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cloudbuild Suite")
}

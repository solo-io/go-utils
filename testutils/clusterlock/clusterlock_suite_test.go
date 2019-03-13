package clusterlock_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClusterlock(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clusterlock Suite")
}

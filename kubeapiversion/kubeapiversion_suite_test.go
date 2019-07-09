package kubeapiversion_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKubeapiversion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kube Api Version Suite")
}

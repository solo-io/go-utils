package kubeapi_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKubeApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kube Api Suite")
}

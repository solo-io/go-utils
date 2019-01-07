package protoutils

import (
	"testing"

	"github.com/solo-io/go-utils/common/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestProtoutil(t *testing.T) {
	RegisterFailHandler(Fail)
	logger.DefaultOut = GinkgoWriter
	RunSpecs(t, "Protoutil Suite")
}

package helminstall_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/helminstall"
)

var _ = Describe("helm install client", func() {
	const (
		namespace = "test-namespace"
	)

	It("can properly set cli env settings with namespace", func() {
		settings := helminstall.NewCLISettings(namespace)
		Expect(settings.Namespace()).To(Equal(namespace))
	})
})

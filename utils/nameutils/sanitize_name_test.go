package nameutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation"

	. "github.com/solo-io/go-utils/utils/nameutils"
)

var _ = Describe("SanitizeName", func() {
	expected := "istio-release-testu4j66c6y-details-istio-routi-59cd38e6f450a2fd"
	It("makes names dns-compliant", func() {
		name := "istio-release-testu4j66c6y-details-istio-routing-testfi4kb3wb-svc-cluster-local"
		actual := SanitizeName(name)
		Expect(len(actual)).To(Equal(63))
		Expect(actual).To(Equal(expected))
		Expect(validation.IsDNS1123Label(actual)).To(HaveLen(0))
	})
})

package certutils_test

import (
	"crypto/x509"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/go-utils/certutils"
	"k8s.io/kubernetes/staging/src/k8s.io/client-go/util/cert"
)

var _ = Describe("GenCert", func() {
	It("successfully generates a cert for the given config", func() {
		certs, err := GenerateSelfSignedCertificate(cert.Config{
			CommonName:   "secure.af",
			Organization: []string{"solo.io"},
			AltNames: cert.AltNames{
				DNSNames: []string{
					"secure.af",
					"secure.af.svc",
					"secure.af.svc.cluster.local",
				},
			},
			Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(certs.CaCertificate)).To(BeNumerically(">", 1))
		Expect(len(certs.ServerCertificate)).To(BeNumerically(">", 1))
		Expect(len(certs.ServerCertKey)).To(BeNumerically(">", 1))
	})
})

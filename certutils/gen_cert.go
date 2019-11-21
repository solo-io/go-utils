package certutils

// imported from https://github.com/solo-io/supergloo/blob/master/pkg/registration/appmesh/tls.go

import (
	"crypto/rand"
	"crypto/rsa"

	"k8s.io/client-go/util/cert"

	"github.com/solo-io/go-utils/errors"
)

type Certificates struct {
	// PEM-encoded CA certificate that has been used to sign the server certificate
	CaCertificate []byte
	// PEM-encoded server certificate
	ServerCertificate []byte
	// PEM-encoded private key that has been used to sign the server certificate
	ServerCertKey []byte
}

// This function generates a self-signed TLS certificate
func GenerateSelfSignedCertificate(config cert.Config) (*Certificates, error) {

	// Generate the CA certificate that will be used to sign the webhook server certificate
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create CA private key")
	}
	caCert, err := cert.NewSelfSignedCACert(cert.Config{CommonName: "supergloo-webhook-cert-ca"}, caPrivateKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create CA certificate")
	}

	// Generate webhook server certificate
	serverCertPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create server cert private key")
	}
	signedServerCert, err := cert.NewSignedCert(config, serverCertPrivateKey, caCert, caPrivateKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create server cert")
	}

	serverCertPrivateKeyPEM, err := cert.MarshalPrivateKeyToPEM(serverCertPrivateKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert server cert private key to PEM")
	}

	return &Certificates{
		CaCertificate:     cert.EncodeCertPEM(caCert),
		ServerCertificate: cert.EncodeCertPEM(signedServerCert),
		ServerCertKey:     serverCertPrivateKeyPEM,
	}, nil
}

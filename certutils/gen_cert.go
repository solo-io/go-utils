package certutils

// imported from https://github.com/solo-io/supergloo/blob/master/pkg/registration/appmesh/tls.go

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math"
	"math/big"
	"time"

	"k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"

	"github.com/pkg/errors"
	"github.com/rotisserie/eris"
)

const (
	certificateBlockType = "CERTIFICATE"
	rsaKeySize           = 2048
	duration365d         = time.Hour * 24 * 365
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
	signedServerCert, err := NewSignedCert(&config, serverCertPrivateKey, caCert, caPrivateKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create server cert")
	}

	serverCertPrivateKeyPEM, err := keyutil.MarshalPrivateKeyToPEM(serverCertPrivateKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert server cert private key to PEM")
	}

	return &Certificates{
		CaCertificate:     EncodeCertPEM(caCert),
		ServerCertificate: EncodeCertPEM(signedServerCert),
		ServerCertKey:     serverCertPrivateKeyPEM,
	}, nil
}

// NewPrivateKey creates an RSA private key
func NewPrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, rsaKeySize)
}

// EncodeCertPEM returns PEM-endcoded certificate data
func EncodeCertPEM(cert *x509.Certificate) []byte {
	block := pem.Block{
		Type:  certificateBlockType,
		Bytes: cert.Raw,
	}
	return pem.EncodeToMemory(&block)
}

// NewSignedCert creates a signed certificate using the given CA certificate and key
func NewSignedCert(cfg *cert.Config, key crypto.Signer, caCert *x509.Certificate, caKey crypto.Signer) (*x509.Certificate, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}
	if len(cfg.CommonName) == 0 {
		return nil, eris.New("must specify a CommonName")
	}
	if len(cfg.Usages) == 0 {
		return nil, eris.New("must specify at least one ExtKeyUsage")
	}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:     cfg.AltNames.DNSNames,
		IPAddresses:  cfg.AltNames.IPs,
		SerialNumber: serial,
		NotBefore:    caCert.NotBefore,
		NotAfter:     time.Now().Add(duration365d).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  cfg.Usages,
	}
	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}

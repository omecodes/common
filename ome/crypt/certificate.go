package crypt

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"time"
)

//CertificateTemplate specs for generating a certificate
type CertificateTemplate struct {
	Organization      string
	Name              string
	Domains           []string
	IPs               []net.IP
	Expiry            time.Duration
	PublicKey         crypto.PublicKey
	SignerPrivateKey  crypto.PrivateKey
	SignerCertificate *x509.Certificate
}

func serialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, _ := rand.Int(rand.Reader, serialNumberLimit)
	return serial
}

func caKeyUsage() x509.KeyUsage {
	return x509.KeyUsageDigitalSignature |
		x509.KeyUsageContentCommitment |
		x509.KeyUsageKeyEncipherment |
		x509.KeyUsageDataEncipherment |
		x509.KeyUsageKeyAgreement |
		x509.KeyUsageCertSign |
		x509.KeyUsageCRLSign
}

func caExtKeyUsage() []x509.ExtKeyUsage {
	return []x509.ExtKeyUsage{x509.ExtKeyUsageAny}
}

func serviceExtKeyUsage() []x509.ExtKeyUsage {
	return []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
}

func serviceKeyUsage() x509.KeyUsage {
	return x509.KeyUsageDigitalSignature |
		x509.KeyUsageContentCommitment |
		x509.KeyUsageKeyEncipherment |
		x509.KeyUsageDataEncipherment |
		x509.KeyUsageKeyAgreement |
		x509.KeyUsageCertSign |
		x509.KeyUsageCRLSign
}

//GenerateCACertificate generates a certificate for a CA
func GenerateCACertificate(t *CertificateTemplate) (*x509.Certificate, error) {
	notBefore := time.Now()
	notAfter := notBefore.Add(t.Expiry)
	template := &x509.Certificate{
		Subject: pkix.Name{
			Organization: []string{t.Organization},
			CommonName:   t.Name,
		},
		SerialNumber:                serialNumber(),
		IsCA:                        true,
		PublicKey:                   t.PublicKey,
		NotBefore:                   notBefore,
		NotAfter:                    notAfter,
		IPAddresses:                 t.IPs,
		DNSNames:                    t.Domains,
		KeyUsage:                    caKeyUsage(),
		ExtKeyUsage:                 caExtKeyUsage(),
		BasicConstraintsValid:       true,
		MaxPathLenZero:              true,
		PermittedDNSDomainsCritical: false,
		PermittedDNSDomains:         nil,
	}

	var (
		pubBytes []byte
		err      error
	)

	switch pk := t.PublicKey.(type) {
	case *ecdsa.PublicKey:
		pubBytes = elliptic.Marshal(elliptic.P521(), pk.X, pk.Y)
	case *rsa.PublicKey:
		pubBytes, err = asn1.Marshal(rsa.PublicKey{
			N: pk.N,
			E: pk.E,
		})
	default:
		err = errors.New("unsupported key type")
	}
	if err != nil {
		return nil, err
	}

	hash := sha1.Sum(pubBytes)
	template.SubjectKeyId = hash[:]

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, t.PublicKey, t.SignerPrivateKey)
	if err != nil {
		return nil, errors.New("")
	}
	return x509.ParseCertificate(certBytes)
}

//GenerateServiceCertificate generates a certificate for a service
func GenerateServiceCertificate(t *CertificateTemplate) (*x509.Certificate, error) {

	notBefore := time.Now()
	notAfter := notBefore.Add(t.Expiry)
	template := &x509.Certificate{
		Subject: pkix.Name{
			Organization: []string{t.Organization},
			CommonName:   t.Name,
		},
		AuthorityKeyId: t.SignerCertificate.SubjectKeyId,
		SerialNumber:   serialNumber(),
		IsCA:           false,
		PublicKey:      t.PublicKey,
		IPAddresses:    t.IPs,
		DNSNames:       t.Domains,
		KeyUsage:       serviceKeyUsage(),
		ExtKeyUsage:    serviceExtKeyUsage(),
		NotBefore:      notBefore,
		NotAfter:       notAfter,
	}

	var (
		pubBytes []byte
		err      error
	)

	switch pk := t.PublicKey.(type) {
	case *ecdsa.PublicKey:
		pubBytes = elliptic.Marshal(elliptic.P521(), pk.X, pk.Y)
	case *rsa.PublicKey:
		pubBytes, err = asn1.Marshal(rsa.PublicKey{
			N: pk.N,
			E: pk.E,
		})
	default:
		err = errors.New("unsupported key type")
	}
	if err != nil {
		return nil, err
	}

	hash := sha1.Sum(pubBytes)
	template.SubjectKeyId = hash[:]

	certBytes, err := x509.CreateCertificate(rand.Reader, template, t.SignerCertificate, t.PublicKey, t.SignerPrivateKey)
	if err != nil {
		return nil, err
	}
	cert := &x509.Certificate{Raw: certBytes}
	return cert, nil
}

//LoadPrivateKey load encrypted private key from "file" and decrypts it
func LoadPrivateKey(password []byte, file string) (crypto.PrivateKey, error) {
	keyBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("bad input")
	}

	if block.Type != "RSA PRIVATE KEY" && block.Type != "ECDSA PRIVATE KEY" {
		return nil, errors.New("key not supported")
	}

	if password != nil && len(password) > 0 {
		keyBytes, err = x509.DecryptPEMBlock(block, password)
		if err != nil {
			return nil, errors.New("bad input")
		}
	} else {
		keyBytes = block.Bytes
	}

	if block.Type == "RSA PRIVATE KEY" {
		return x509.ParsePKCS1PrivateKey(keyBytes)
	}

	return x509.ParseECPrivateKey(keyBytes)
}

//StorePrivateKey encrypts the private key and save it in "file"
func StorePrivateKey(key crypto.PrivateKey, password []byte, file string) error {
	var block *pem.Block
	var err error

	if rp, ok := key.(*rsa.PrivateKey); ok {
		block = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rp)}

	} else if ep, ok := key.(*ecdsa.PrivateKey); ok {
		privateKeyBytes, err := x509.MarshalECPrivateKey(ep)
		if err != nil {
			return err
		}
		block = &pem.Block{Type: "ECDSA PRIVATE KEY", Bytes: privateKeyBytes}

	} else {
		return errors.New("not supported")
	}

	if password != nil && len(password) > 0 {
		block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, password, x509.PEMCipherAES256)
		if err != nil {
			return err
		}
	}
	return ioutil.WriteFile(file, pem.EncodeToMemory(block), 0600)
}

//LoadCertificate load file and decode it into a x509.Certificate
func LoadCertificate(file string) (*x509.Certificate, error) {
	certBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(certBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate from %s file", file)
	}
	return x509.ParseCertificate(block.Bytes)
}

//StoreCertificate encode certificate and store the result in "file"
func StoreCertificate(cert *x509.Certificate, file string, perm os.FileMode) error {
	buff := bytes.NewBuffer([]byte{})
	err := pem.Encode(buff, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, buff.Bytes(), os.ModePerm)
}

func PEMEncodeCertificate(cert *x509.Certificate) ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})
	err := pem.Encode(buff, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func PEMDecodeCertificate(pemBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("could not decode input bytes")
	}
	return x509.ParseCertificate(block.Bytes)
}

func PEMEncodeKey(key crypto.PrivateKey) ([]byte, error) {
	var block *pem.Block

	if rp, ok := key.(*rsa.PrivateKey); ok {
		block = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rp)}

	} else if ep, ok := key.(*ecdsa.PrivateKey); ok {
		privateKeyBytes, err := x509.MarshalECPrivateKey(ep)
		if err != nil {
			return nil, err
		}
		block = &pem.Block{Type: "ECDSA PRIVATE KEY", Bytes: privateKeyBytes}
	} else {
		return nil, errors.New("not supported")
	}
	return pem.EncodeToMemory(block), nil
}

func PEMEncodePublicKey(k crypto.PublicKey) ([]byte, error) {
	var block *pem.Block
	b, err := x509.MarshalPKIXPublicKey(k)
	if err != nil {
		return nil, err
	}

	block = &pem.Block{Bytes: b}

	switch k.(type) {
	case *ecdsa.PublicKey:
		block.Type = "ECDSA PUBLIC KEY"
	case *rsa.PublicKey:
		block.Type = "RSA PUBLIC KEY"
	default:
		block.Type = "PUBLIC KEY"
	}

	return pem.EncodeToMemory(block), nil
}

func PEMDecodePublicKey(pemBytes []byte) (interface{}, string, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, "", errors.New("could not decode PEM bytes")
	}

	k, err := x509.ParsePKIXPublicKey(block.Bytes)
	return k, block.Type, err
}

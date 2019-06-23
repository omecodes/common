package crypto2

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"github.com/zoenion/common/errors"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

//Template specs for generating a certificate
type Template struct {
	Organization      string
	Name              string
	Domains           []string
	IPs               []net.IP
	Expiry            time.Duration
	PublicKey         crypto.PublicKey
	SignerPrivateKey  crypto.PrivateKey
	SignerCertificate *x509.Certificate
}

func localIPs() []net.IP {
	ips := []net.IP{}
	ifaces, err := net.InterfaceAddrs()
	if err == nil {
		for i := range ifaces {
			addr := ifaces[i]
			ips = append(ips, net.ParseIP(strings.Split(addr.String(), "/")[0]))
		}
	}
	return ips
}

func serialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, _ := rand.Int(rand.Reader, serialNumberLimit)
	return serial
}

func caKeyUsage() x509.KeyUsage {
	return x509.KeyUsageCertSign | x509.KeyUsageCRLSign
}

func caExtKeyUsage() []x509.ExtKeyUsage {
	return []x509.ExtKeyUsage{x509.ExtKeyUsageAny}
}

func serviceExtKeyUsage() []x509.ExtKeyUsage {
	return []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
}

func serviceKeyUsage() x509.KeyUsage {
	return x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment
}

//GenerateCACertificate generates a certificate for a CA
func GenerateCACertificate(t *Template) (*x509.Certificate, error) {
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

	pubBytes, err := asn1.Marshal(rsa.PublicKey{
		N: (t.PublicKey.(*rsa.PublicKey)).N,
		E: (t.PublicKey.(*rsa.PublicKey)).E,
	})
	if err == nil {
		hash := sha1.Sum(pubBytes)
		template.SubjectKeyId = hash[:]
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, t.PublicKey, t.SignerPrivateKey)
	if err != nil {
		return nil, errors.Detailed(errors.BadInput, "failed to create CA certificate: "+err.Error())
	}
	return x509.ParseCertificate(certBytes)
}

//GenerateServiceCertificate generates a certificate for a service
func GenerateServiceCertificate(t *Template) (*x509.Certificate, error) {

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

	if pub, ok := t.PublicKey.(*rsa.PublicKey); ok {
		pubBytes, err := asn1.Marshal(rsa.PublicKey{
			N: pub.N,
			E: pub.E,
		})
		if err == nil {
			hash := sha1.Sum(pubBytes)
			template.SubjectKeyId = hash[:]
		}
	}

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
		return nil, errors.Detailed(errors.BadInput, "failed to decode the private key")
	}

	if block.Type != "RSA PRIVATE KEY" && block.Type != "ECDSA PRIVATE KEY" {
		return nil, errors.Detailed(errors.HttpNotImplemented, "key type not supported")
	}

	if password != nil && len(password) > 0 {
		keyBytes, err = x509.DecryptPEMBlock(block, password)
		if err != nil {
			return nil, errors.Detailed(errors.BadInput, "Failed to decrypt CA key: "+err.Error())
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
		bytes, err := x509.MarshalECPrivateKey(ep)
		if err != nil {
			return err
		}
		block = &pem.Block{Type: "ECDSA PRIVATE KEY", Bytes: bytes}

	} else {
		return errors.Detailed(errors.HttpNotImplemented, "key type is not supported")
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

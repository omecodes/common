package netx

import (
	"crypto/tls"
	"crypto/x509"
	crypto2 "github.com/omecodes/common/security/crypto"
	"net"
)

type ListenOptions struct {
	Trust        bool
	CertFilename string
	KeyFilename  string
	TLS          *tls.Config
	Secure       bool
	KeyPassword  []byte
}

// ListenOption enriches listen options object
type ListenOption func(opts *ListenOptions)

// TrustSSL when set to true when the certificate is self-signed
func TrustSSL(trust bool) ListenOption {
	return func(opts *ListenOptions) {
		opts.Trust = trust
	}
}

// TLS tls config for listen to secure connections
func TLS(tc *tls.Config) ListenOption {
	return func(opts *ListenOptions) {
		opts.Secure = tc != nil
		opts.TLS = tc
		opts.CertFilename = ""
		opts.KeyFilename = ""
	}
}

// Secure specify certificate and key filenames for tls config
func Secure(certFilename, keyFilename string) ListenOption {
	return func(opts *ListenOptions) {
		opts.Secure = certFilename != "" && keyFilename != ""
		opts.TLS = nil
		opts.CertFilename = certFilename
		opts.KeyFilename = keyFilename
	}
}

// KeyPassword passed if the key filename is protected
func KeyPassword(password []byte) ListenOption {
	return func(opts *ListenOptions) {
		opts.KeyPassword = password
	}
}

// Listen listen to tcp connections
func Listen(address string, opts ...ListenOption) (net.Listener, error) {
	var lopts ListenOptions
	for _, opt := range opts {
		opt(&lopts)
	}

	if address == "" {
		address = ":"
	}

	if lopts.CertFilename != "" && lopts.KeyFilename != "" {

		cert, err := crypto2.LoadCertificate(lopts.CertFilename)
		if err != nil {
			return nil, err
		}

		key, err := crypto2.LoadPrivateKey(lopts.KeyPassword, lopts.KeyFilename)
		if err != nil {
			return nil, err
		}

		tc := &tls.Config{
			Certificates: []tls.Certificate{
				{
					Certificate: [][]byte{cert.Raw},
					PrivateKey:  key,
				},
			},
		}

		if lopts.Trust {
			pool := x509.NewCertPool()
			pool.AddCert(cert)
			tc.ClientCAs = pool
		}

		return tls.Listen("tcp", address, tc)

	} else if lopts.TLS != nil {
		return tls.Listen("tcp", address, lopts.TLS)
	} else {
		return net.Listen("tcp", address)
	}
}

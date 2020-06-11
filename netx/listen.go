package netx

import (
	"crypto/tls"
	"crypto/x509"
	crypto2 "github.com/omecodes/common/crypto"
	"net"
)

type listenOptions struct {
	selfSigned   bool
	certFilename string
	keyFilename  string
	tlsConfig    *tls.Config
	keyPassword  []byte
}

// ListenOption enriches listen options object
type ListenOption func(opts *listenOptions)

// TrustSSL when set to true when the certificate is self-signed
func TrustSSL(trust bool) ListenOption {
	return func(opts *listenOptions) {
		opts.selfSigned = trust
	}
}

// TLS tls config for listen to secure connections
func TLS(tc *tls.Config) ListenOption {
	return func(opts *listenOptions) {
		opts.tlsConfig = tc
		opts.certFilename = ""
		opts.keyFilename = ""
	}
}

// Secure specify certificate and key filenames for tls config
func Secure(certFilename, keyFilename string) ListenOption {
	return func(opts *listenOptions) {
		opts.tlsConfig = nil
		opts.certFilename = certFilename
		opts.keyFilename = keyFilename
	}
}

// KeyPassword passed if the key filename is protected
func KeyPassword(password []byte) ListenOption {
	return func(opts *listenOptions) {
		opts.keyPassword = password
	}
}

// Listen listen to tcp connections
func Listen(address string, opts ...ListenOption) (net.Listener, error) {
	var lopts listenOptions
	for _, opt := range opts {
		opt(&lopts)
	}

	if address == "" {
		address = ":"
	}

	if lopts.certFilename != "" && lopts.keyFilename != "" {

		cert, err := crypto2.LoadCertificate(lopts.certFilename)
		if err != nil {
			return nil, err
		}

		key, err := crypto2.LoadPrivateKey(lopts.keyPassword, lopts.keyFilename)
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

		if lopts.selfSigned {
			pool := x509.NewCertPool()
			pool.AddCert(cert)
			tc.ClientCAs = pool
		}

		return tls.Listen("tcp", address, tc)

	} else if lopts.tlsConfig != nil {
		return tls.Listen("tcp", address, lopts.tlsConfig)
	} else {
		return net.Listen("tcp", address)
	}
}

package net

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func GRPCMutualTlsDial(address string, authorityCert, clientCert *x509.Certificate, clientKey crypto.PrivateKey) (*grpc.ClientConn, error) {
	CAPool := x509.NewCertPool()
	CAPool.AddCert(authorityCert)

	tlsCert := tls.Certificate{
		Certificate: [][]byte{clientCert.Raw},
		PrivateKey:  clientKey,
	}

	tlsCfg := &tls.Config{
		RootCAs:      CAPool,
		Certificates: []tls.Certificate{tlsCert},
	}

	return grpc.Dial(address, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
}

func GRPCMutualTlsDialWithCredentials(address string, authorityCert, clientCert *x509.Certificate, clientKey crypto.PrivateKey, cred credentials.PerRPCCredentials) (*grpc.ClientConn, error) {
	CAPool := x509.NewCertPool()
	CAPool.AddCert(authorityCert)

	tlsCert := tls.Certificate{
		Certificate: [][]byte{clientCert.Raw},
		PrivateKey:  clientKey,
	}

	tlsCfg := &tls.Config{
		RootCAs:      CAPool,
		Certificates: []tls.Certificate{tlsCert},
	}

	return grpc.Dial(address, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)), grpc.WithPerRPCCredentials(cred))
}

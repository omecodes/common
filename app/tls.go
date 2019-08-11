package app

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/zoenion/common/auth"
	crypto2 "github.com/zoenion/common/crypto"
	"github.com/zoenion/common/errors"
	"github.com/zoenion/common/futils"
	authoritypb "github.com/zoenion/common/proto/authority"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func GRPCAuthorityTransportCredentials(a *Vars) (credentials.TransportCredentials, error) {
	if a.gRPCAuthorityCredentials != nil {
		return a.gRPCAuthorityCredentials, nil
	}

	if a.AuthorityCertPath == "" {
		return nil, errors.BadInput
	}
	if !futils.FileExists(a.AuthorityCertPath) {
		return nil, errors.NotFound
	}
	var err error
	a.gRPCAuthorityCredentials, err = credentials.NewClientTLSFromFile(a.AuthorityCertPath, "")
	return a.gRPCAuthorityCredentials, err
}

func LoadTLS(a *Vars) (*tls.Config, *tls.Config, error) {
	if a.tlsClient != nil && a.tlsServer != nil {
		return a.tlsServer, a.tlsClient, nil
	}

	if a.AuthorityCertPath == "" {
		return nil, nil, errors.BadInput
	}
	if !futils.FileExists(a.AuthorityCertPath) {
		return nil, nil, errors.NotFound
	}
	authorityCert, err := crypto2.LoadCertificate(a.AuthorityCertPath)
	if err != nil {
		return nil, nil, fmt.Errorf("could not load authority certificate: %s", err)
	}

	name := strcase.ToSnake(a.Name)
	certFilename := filepath.Join(a.Dir, fmt.Sprintf("%s.crt", name))
	keyFilename := filepath.Join(a.Dir, fmt.Sprintf("%s.key", name))

	var (
		serviceKey  crypto.PrivateKey
		serviceCert *x509.Certificate
	)

	shouldGenerateNewPair := !futils.FileExists(certFilename) || !futils.FileExists(keyFilename)
	if !shouldGenerateNewPair {
		serviceKey, err = crypto2.LoadPrivateKey([]byte{}, keyFilename)
		if err != nil {
			return nil, nil, fmt.Errorf("could not load private key: %s", err)
		}
		serviceCert, err = crypto2.LoadCertificate(certFilename)
		if err != nil {
			return nil, nil, fmt.Errorf("could not load certificate: %s", err)
		}
	}

	CAPool := x509.NewCertPool()
	CAPool.AddCert(authorityCert)

	if serviceCert != nil {
		_, err = serviceCert.Verify(x509.VerifyOptions{Roots: CAPool})
		if err != nil || time.Now().After(serviceCert.NotAfter) || time.Now().Before(serviceCert.NotBefore) {
			shouldGenerateNewPair = true
		}
	}

	if shouldGenerateNewPair {
		serviceKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		pub := serviceKey.(*ecdsa.PrivateKey).PublicKey

		creds, err := GRPCAuthorityTransportCredentials(a)
		if err != nil {
			return nil, nil, err
		}

		if a.authorityGRPCAuthentication == nil {
			parts := strings.Split(a.AuthorityCredentials, ":")
			a.authorityGRPCAuthentication = auth.NewGRPCBasicAuthentication(parts[0], parts[1])
		}

		conn, err := grpc.Dial(a.GRPCAuthorityAddress, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(a.authorityGRPCAuthentication))
		client := authoritypb.NewAuthorityServiceClient(conn)
		rsp, err := client.SignCertificate(context.Background(), &authoritypb.SignCertificateRequest{
			Template: &authoritypb.CertificateTemplate{
				Domains:     []string{a.Domain},
				Addresses:   []string{a.IP},
				ServiceName: strcase.ToDelimited(a.Name, '.'),
				PublicKey:   elliptic.Marshal(elliptic.P256(), pub.X, pub.Y),
			},
		})
		if err != nil {
			return nil, nil, err
		}

		serviceCert, err = x509.ParseCertificate(rsp.RawCertificate)
		if err != nil {
			return nil, nil, err
		}

		err = crypto2.StoreCertificate(serviceCert, certFilename, os.ModePerm)
		if err != nil {
			return nil, nil, err
		}

		err = crypto2.StorePrivateKey(serviceKey, nil, keyFilename)
		if err != nil {
			return nil, nil, err
		}
	}

	tlsCert := tls.Certificate{
		Certificate: [][]byte{serviceCert.Raw},
		PrivateKey:  serviceKey.(*ecdsa.PrivateKey),
	}

	a.tlsServer = &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ClientCAs:    CAPool,
		ClientAuth:   tls.VerifyClientCertIfGiven,
		ServerName:   a.Domain,
	}
	a.tlsClient = &tls.Config{
		RootCAs:      CAPool,
		Certificates: []tls.Certificate{tlsCert},
	}
	return a.tlsServer, a.tlsClient, nil
}

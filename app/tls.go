package app

import (
	"context"
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

func authorityTransportCredentials(a *Vars) (credentials.TransportCredentials, error) {
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

func loadSignedKeyPair(v *Vars) error {
	if v.serviceCert != nil && v.serviceKey != nil {
		return nil
	}

	if v.AuthorityCertPath == "" {
		return errors.BadInput
	}

	v.AuthorityCertPath, _ = filepath.Abs(v.AuthorityCertPath)
	if !futils.FileExists(v.AuthorityCertPath) {
		return errors.NotFound
	}
	authorityCert, err := crypto2.LoadCertificate(v.AuthorityCertPath)
	if err != nil {
		return fmt.Errorf("could not load authority certificate: %s", err)
	}

	v.authorityCert = authorityCert

	name := strcase.ToSnake(v.Name)
	certFilename := filepath.Join(v.Dir, fmt.Sprintf("%s.crt", name))
	keyFilename := filepath.Join(v.Dir, fmt.Sprintf("%s.key", name))

	shouldGenerateNewPair := !futils.FileExists(certFilename) || !futils.FileExists(keyFilename)
	if !shouldGenerateNewPair {
		v.serviceKey, err = crypto2.LoadPrivateKey([]byte{}, keyFilename)
		if err != nil {
			return fmt.Errorf("could not load private key: %s", err)
		}
		v.serviceCert, err = crypto2.LoadCertificate(certFilename)
		if err != nil {
			return fmt.Errorf("could not load certificate: %s", err)
		}
	}

	CAPool := x509.NewCertPool()
	CAPool.AddCert(authorityCert)

	if v.serviceCert != nil {
		_, err = v.serviceCert.Verify(x509.VerifyOptions{Roots: CAPool})
		if err != nil || time.Now().After(v.serviceCert.NotAfter) || time.Now().Before(v.serviceCert.NotBefore) {
			shouldGenerateNewPair = true
		}
	}

	if shouldGenerateNewPair {
		v.serviceKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return fmt.Errorf("could not generate key pair: %s", err)
		}
		pub := v.serviceKey.(*ecdsa.PrivateKey).PublicKey

		v.gRPCAuthorityCredentials, err = authorityTransportCredentials(v)
		if err != nil {
			return err
		}

		if v.authorityGRPCAuthentication == nil {
			parts := strings.Split(v.AuthorityCredentials, ":")
			v.authorityGRPCAuthentication = auth.NewGRPCBasicAuthentication(parts[0], parts[1])
		}

		conn, err := grpc.Dial(v.GRPCAuthorityAddress, grpc.WithTransportCredentials(v.gRPCAuthorityCredentials), grpc.WithPerRPCCredentials(v.authorityGRPCAuthentication))
		client := authoritypb.NewAuthorityServiceClient(conn)
		rsp, err := client.SignCertificate(context.Background(), &authoritypb.SignCertificateRequest{
			Template: &authoritypb.CertificateTemplate{
				Domains:     []string{v.Domain},
				Addresses:   []string{v.IP},
				ServiceName: strcase.ToDelimited(v.Name, '.'),
				PublicKey:   elliptic.Marshal(elliptic.P256(), pub.X, pub.Y),
			},
		})
		if err != nil {
			return fmt.Errorf("could not sign certificate: %s", err)
		}

		v.serviceCert, err = x509.ParseCertificate(rsp.RawCertificate)
		if err != nil {
			return err
		}

		_ = crypto2.StoreCertificate(v.serviceCert, certFilename, os.ModePerm)
		_ = crypto2.StorePrivateKey(v.serviceKey, nil, keyFilename)
	}
	return nil
}

func ServerMutualTLS(v *Vars) *tls.Config {
	if v.serviceKey == nil || v.serviceCert == nil || v.authorityCert == nil {
		return nil
	}

	CAPool := x509.NewCertPool()
	CAPool.AddCert(v.authorityCert)

	tlsCert := tls.Certificate{
		Certificate: [][]byte{v.serviceCert.Raw},
		PrivateKey:  v.serviceKey,
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ClientCAs:    CAPool,
		ClientAuth:   tls.RequestClientCert,
		ServerName:   v.Domain,
	}
}

func ClientMutualTLS(v *Vars) *tls.Config {
	if v.serviceKey == nil || v.serviceCert == nil || v.authorityCert == nil {
		return nil
	}

	CAPool := x509.NewCertPool()
	CAPool.AddCert(v.authorityCert)

	tlsCert := tls.Certificate{
		Certificate: [][]byte{v.serviceCert.Raw},
		PrivateKey:  v.serviceKey.(*ecdsa.PrivateKey),
	}

	return &tls.Config{
		RootCAs:      CAPool,
		Certificates: []tls.Certificate{tlsCert},
	}
}

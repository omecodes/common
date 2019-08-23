package service

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
	capb "github.com/zoenion/common/proto/ca"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Box struct {
	params BoxParams

	gateway                           *gateway
	registry                          *SyncedRegistry
	authorityCert                     *x509.Certificate
	authorityClientAuthentication     credentials.PerRPCCredentials
	authorityGRPCTransportCredentials credentials.TransportCredentials
	registryCert                      *x509.Certificate
	serviceCert                       *x509.Certificate
	serviceKey                        crypto.PrivateKey
}

func (box *Box) Name() string {
	return box.params.Name
}

func (box *Box) Dir() string {
	return box.params.Dir
}

func (box *Box) Ip() string {
	return box.params.Ip
}

func (box *Box) RegistryCert() *x509.Certificate {
	return box.registryCert
}

func (box *Box) Registry() *SyncedRegistry {
	return box.registry
}

func (box *Box) AuthorityCert() *x509.Certificate {
	return box.authorityCert
}

func (box *Box) AuthorityClientAuthentication() credentials.PerRPCCredentials {
	return box.authorityClientAuthentication
}

func (box *Box) AuthorityGRPCTransportCredentials() credentials.TransportCredentials {
	return box.authorityGRPCTransportCredentials
}

func (box *Box) ServiceCert() *x509.Certificate {
	return box.serviceCert
}

func (box *Box) ServiceKey() crypto.PrivateKey {
	return box.serviceKey
}

func (box *Box) GRPCAddress() string {
	addr := box.params.Domain
	if addr == "" {
		addr = box.params.Ip
	}
	return fmt.Sprintf("%s:%s", addr, box.params.GatewayGRPCPort)
}

func (box *Box) HTTPAddress() string {
	addr := box.params.Domain
	if addr == "" {
		addr = box.params.Ip
	}
	return fmt.Sprintf("%s:%s", addr, box.params.GatewayHTTPPort)
}

func (box *Box) validateParams() error {
	if box.params.Name == "" {
		return errors.New("command line: --name flags is required")
	}

	if box.params.Domain == "" && box.params.Ip == "" {
		return errors.New("command line: one or both --domain and --ip flags must be passed")
	}

	if box.params.Dir == "" {
		d := getDir()
		box.params.Dir = d.Path()
		if err := d.Create(); err != nil {
			log.Printf("command line: could not create %s. Might not be writeable\n", box.Dir)
			return err
		}
	} else {
		var err error
		box.params.Dir, err = filepath.Abs(box.params.Dir)
		if err != nil {
			log.Printf("command line: could not find %s\n", box.Dir)
			return err
		}
	}

	if box.params.CaGRPC != "" || box.params.CaCertPath != "" || box.params.CaCredentials != "" {
		if box.params.CaGRPC == "" || box.params.CaCertPath == "" || box.params.CaCredentials == "" {
			return fmt.Errorf("command line: --a-grpc must always be provided with --a-cert and --a-cred")
		}
	}

	if box.params.RegistryAddress != "" || box.params.Namespace != "" {
		if box.params.RegistryAddress == "" || box.params.Namespace == "" {
			return errors.New("command line: --namespace must always be provided with --registryAddress")
		}
	}

	if box.params.CaCertPath != "" || box.params.KeyPath != "" {
		if box.params.CaCertPath == "" || box.params.KeyPath == "" {
			return errors.New("command line: --cert must always be provided with --key")
		}
	}

	return nil
}

func (box *Box) loadTools() error {
	var err error

	if box.params.CertificatePath != "" {
		box.serviceCert, err = crypto2.LoadCertificate(box.params.CaCertPath)
		if err != nil {
			return fmt.Errorf("could not load service certificate: %s", err)
		}

		box.serviceKey, err = crypto2.LoadPrivateKey(nil, box.params.KeyPath)
		if err != nil {
			return fmt.Errorf("could not load service private key: %s", err)
		}
	}

	if box.params.CaGRPC != "" {
		box.authorityCert, err = crypto2.LoadCertificate(box.params.CaCertPath)
		if err != nil {
			return fmt.Errorf("could not load authority certificate: %s", err)
		}

		box.authorityGRPCTransportCredentials, err = credentials.NewClientTLSFromFile(box.params.CaCertPath, "")
		if err != nil {
			return fmt.Errorf("could not create authority client credentials: %s", box.params.CaCertPath)
		}

		parts := strings.Split(box.params.CaCredentials, ":")
		box.authorityClientAuthentication = auth.NewGRPCBasicAuthentication(parts[0], parts[1])

		err = box.loadSignedKeyPair()
		if err != nil {
			return err
		}
	}

	if box.params.RegistryAddress != "" {
		if box.params.RegistrySecure {
			box.registry = NewSyncRegistry(box.params.RegistryAddress, box.clientMutualTLS())
		} else {
			box.registry = NewSyncRegistry(box.params.RegistryAddress, nil)
		}
	}
	return nil
}

func (box *Box) loadSignedKeyPair() error {
	if box.serviceCert != nil && box.serviceKey != nil {
		return nil
	}

	if box.params.CaCertPath == "" {
		return errors.BadInput
	}

	box.params.CaCertPath, _ = filepath.Abs(box.params.CaCertPath)
	if !futils.FileExists(box.params.CaCertPath) {
		return errors.NotFound
	}
	authorityCert, err := crypto2.LoadCertificate(box.params.CaCertPath)
	if err != nil {
		return fmt.Errorf("could not load authority certificate: %s", err)
	}

	box.authorityCert = authorityCert

	name := strcase.ToSnake(box.params.Name)
	certFilename := filepath.Join(box.params.Dir, fmt.Sprintf("%s.crt", name))
	keyFilename := filepath.Join(box.params.Dir, fmt.Sprintf("%s.key", name))

	shouldGenerateNewPair := !futils.FileExists(certFilename) || !futils.FileExists(keyFilename)
	if !shouldGenerateNewPair {
		box.serviceKey, err = crypto2.LoadPrivateKey([]byte{}, keyFilename)
		if err != nil {
			return fmt.Errorf("could not load private key: %s", err)
		}
		box.serviceCert, err = crypto2.LoadCertificate(certFilename)
		if err != nil {
			return fmt.Errorf("could not load certificate: %s", err)
		}
	}

	CAPool := x509.NewCertPool()
	CAPool.AddCert(authorityCert)

	if box.serviceCert != nil {
		_, err = box.serviceCert.Verify(x509.VerifyOptions{Roots: CAPool})
		if err != nil || time.Now().After(box.serviceCert.NotAfter) || time.Now().Before(box.serviceCert.NotBefore) {
			shouldGenerateNewPair = true
		}
	}

	if shouldGenerateNewPair {
		box.serviceKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return fmt.Errorf("could not generate key pair: %s", err)
		}
		pub := box.serviceKey.(*ecdsa.PrivateKey).PublicKey

		if box.authorityClientAuthentication == nil {
			parts := strings.Split(box.params.CaCredentials, ":")
			box.authorityClientAuthentication = auth.NewGRPCBasicAuthentication(parts[0], parts[1])
		}

		conn, err := grpc.Dial(box.params.CaGRPC, grpc.WithTransportCredentials(box.authorityGRPCTransportCredentials), grpc.WithPerRPCCredentials(box.authorityClientAuthentication))
		client := capb.NewAuthorityServiceClient(conn)
		rsp, err := client.SignCertificate(context.Background(), &capb.SignCertificateRequest{
			Template: &capb.CertificateTemplate{
				Domains:     []string{box.params.Domain},
				Addresses:   []string{box.params.Ip},
				ServiceName: strcase.ToDelimited(box.params.Name, '.'),
				PublicKey:   elliptic.Marshal(elliptic.P256(), pub.X, pub.Y),
			},
		})
		if err != nil {
			return fmt.Errorf("could not sign certificate: %s", err)
		}

		box.serviceCert, err = x509.ParseCertificate(rsp.RawCertificate)
		if err != nil {
			return err
		}

		_ = crypto2.StoreCertificate(box.serviceCert, certFilename, os.ModePerm)
		_ = crypto2.StorePrivateKey(box.serviceKey, nil, keyFilename)
	}
	return nil
}

func (box *Box) serverMutualTLS() *tls.Config {
	if box.serviceKey == nil || box.serviceCert == nil || box.authorityCert == nil {
		return nil
	}

	CAPool := x509.NewCertPool()
	CAPool.AddCert(box.authorityCert)

	tlsCert := tls.Certificate{
		Certificate: [][]byte{box.serviceCert.Raw},
		PrivateKey:  box.serviceKey,
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ClientCAs:    CAPool,
		ClientAuth:   tls.VerifyClientCertIfGiven,
		ServerName:   box.params.Domain,
	}
}

func (box *Box) clientMutualTLS() *tls.Config {
	if box.serviceKey == nil || box.serviceCert == nil || box.authorityCert == nil {
		return nil
	}

	CAPool := x509.NewCertPool()
	CAPool.AddCert(box.authorityCert)

	tlsCert := tls.Certificate{
		Certificate: [][]byte{box.serviceCert.Raw},
		PrivateKey:  box.serviceKey.(*ecdsa.PrivateKey),
	}

	return &tls.Config{
		RootCAs:      CAPool,
		Certificates: []tls.Certificate{tlsCert},
	}
}

func (box *Box) start(cfg *BoxConfigs) error {

	if cfg.Web.Tls == nil {
		box.gateway.web.Tls = box.serverMutualTLS()
	}

	if cfg.Grpc.Tls == nil {
		box.serverMutualTLS()
	}

	box.gateway = &gateway{
		name:        box.params.Name,
		gRPC:        cfg.Grpc,
		web:         cfg.Web,
		gRPCAddress: box.GRPCAddress(),
		httpAddress: box.HTTPAddress(),
	}

	return box.gateway.start()
}

func (box *Box) stop() {
	if box.gateway != nil {
		box.gateway.stop()
	}
}

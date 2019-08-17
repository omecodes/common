package service

import (
	"crypto"
	"crypto/x509"
	"fmt"
	servicepb "github.com/zoenion/common/proto/service"
	"google.golang.org/grpc/credentials"
)

type loaded struct {
	registryClient                    servicepb.RegistryClient
	authorityCert                     *x509.Certificate
	authorityClientAuthentication     credentials.PerRPCCredentials
	authorityGRPCTransportCredentials credentials.TransportCredentials
	serviceCert                       *x509.Certificate
	serviceKey                        crypto.PrivateKey
}

type Vars struct {
	Name            string
	Dir             string
	Domain          string
	IP              string
	CertificatePath string
	KeyPath         string

	Registry         string
	Namespace        string
	RegistryID       string
	RegistryCertPath string

	AuthorityGRPC        string
	AuthorityCertPath    string
	AuthorityCredentials string

	GatewayGRPCPort string
	GatewayHTTPPort string

	loaded
}

type ConfigVars struct {
	Dir  string
	Name string
}

func (v *Vars) GRPCAddress() string {
	addr := v.Domain
	if addr == "" {
		addr = v.IP
	}
	return fmt.Sprintf("%s:%s", addr, v.GatewayGRPCPort)
}

func (v *Vars) HTTPAddress() string {
	addr := v.Domain
	if addr == "" {
		addr = v.IP
	}
	return fmt.Sprintf("%s:%s", addr, v.GatewayHTTPPort)
}

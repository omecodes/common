package service

import (
	"crypto"
	"crypto/x509"
	"fmt"
	servicepb "github.com/zoenion/common/proto/service"
	"google.golang.org/grpc/credentials"
)

type loaded struct {
	registry                          *servicepb.SyncedRegistry
	authorityCert                     *x509.Certificate
	authorityClientAuthentication     credentials.PerRPCCredentials
	authorityGRPCTransportCredentials credentials.TransportCredentials
	registryCert                      *x509.Certificate
	serviceCert                       *x509.Certificate
	serviceKey                        crypto.PrivateKey
}

type ConfigVars struct {
	dir  string
	name string
}

func (v *ConfigVars) Name() string {
	return v.name
}

func (v *ConfigVars) Dir() string {
	return v.dir
}

type Vars struct {
	name            string
	dir             string
	domain          string
	ip              string
	certificatePath string
	keyPath         string

	registryAddress string
	registrySecure  bool
	namespace       string
	registryID      string

	authorityGRPC        string
	authorityCertPath    string
	authorityCredentials string

	gatewayGRPCPort string
	gatewayHTTPPort string

	loaded
}

func (v *Vars) Name() string {
	return v.name
}

func (v *Vars) Dir() string {
	return v.dir
}

func (v *Vars) Ip() string {
	return v.ip
}

func (v *Vars) RegistryCert() *x509.Certificate {
	return v.registryCert
}

func (v *Vars) Registry() *servicepb.SyncedRegistry {
	return v.registry
}

func (v *Vars) AuthorityCert() *x509.Certificate {
	return v.authorityCert
}

func (v *Vars) AuthorityClientAuthentication() credentials.PerRPCCredentials {
	return v.authorityClientAuthentication
}

func (v *Vars) AuthorityGRPCTransportCredentials() credentials.TransportCredentials {
	return v.authorityGRPCTransportCredentials
}

func (v *Vars) ServiceCert() *x509.Certificate {
	return v.serviceCert
}

func (v *Vars) ServiceKey() crypto.PrivateKey {
	return v.serviceKey
}

func (v *Vars) GRPCAddress() string {
	addr := v.domain
	if addr == "" {
		addr = v.ip
	}
	return fmt.Sprintf("%s:%s", addr, v.gatewayGRPCPort)
}

func (v *Vars) HTTPAddress() string {
	addr := v.domain
	if addr == "" {
		addr = v.ip
	}
	return fmt.Sprintf("%s:%s", addr, v.gatewayHTTPPort)
}

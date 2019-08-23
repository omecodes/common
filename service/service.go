package service

import (
	"crypto/x509"
	servicepb "github.com/zoenion/common/proto/service"
)

const (
	CmdFlagAuthority      = "a-grpc"
	CmdFlagIP             = "ip"
	CmdFlagName           = "name"
	CmdFlagDir            = "dir"
	CmdFlagDomain         = "domain"
	CmdFlagCert           = "cert"
	CmdFlagKey            = "key"
	CmdFlagNamespace      = "ns"
	CmdFlagAuthorityCert  = "a-cert"
	CmdFlagAuthorityCred  = "a-cred"
	CmdFlagRegistry       = "reg"
	CmdFlagRegistrySecure = "reg-secure"
	CmdFlagGRPC           = "grpc"
	CmdFlagHTTP           = "http"
	MetaCertificate       = "certificate"
)

type ContextKey string

const (
	User            = ContextKey("user")
	Caller          = ContextKey("caller")
	Token           = ContextKey("token")
	PeerIp          = ContextKey("peer-ip")
	PeerCertificate = ContextKey("peer-certificate")
)

const (
	ProtocolGRPC = "gRPC"
	ProtocolHTTP = "Web"
)

const (
	AuthBearer  = "Bearer "
	AuthGateway = "gateway "
)

type ConnectionInfo struct {
	Address     string
	Certificate *x509.Certificate
}

type Service interface {
	Type() servicepb.Type
	Configure(name, dir string) error
	Configs(name, dir string) (*BoxConfigs, error)
	AfterStart()
	AfterStop()
}

package service

import "github.com/zoenion/common/service/pb"

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
)

type ContextKey string

const (
	User            = ContextKey("user")
	Caller          = ContextKey("caller")
	Token           = ContextKey("token")
	PeerIp          = ContextKey("peer-ip")
	PeerCertificate = ContextKey("peer-certificate")
)

type Service interface {
	Type() pb.Type
	Configure(name, dir string) error
	Init(name, dir string) (*BoxData, error)
}

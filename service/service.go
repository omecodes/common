package service

import (
	"context"
	"github.com/zoenion/common/service/pb"
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
)

type BoxData struct {
	Type    pb.Type
	Web     *Web
	Grpc    *Grpc
	Options []Option
}

type Service interface {
	Configure(name, dir string) error
	Init(ctx context.Context) (*BoxData, error)
}

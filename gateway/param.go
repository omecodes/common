package gateway

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	http_helper "github.com/zoenion/common/http-helper"
	"google.golang.org/grpc"
)

type HTTP struct {
	Address        string
	WireGRPCFunc   WireEndpointFunc
	MiddlewareList []http_helper.HttpMiddleware
}

type GRPC struct {
	Address             string
	RegisterHandlerFunc func(*grpc.Server)
}

type Config struct {
	Name string
	HTTP *HTTP
	GRPC *GRPC
}

type WireEndpointFunc func(ctx context.Context, serveMux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error

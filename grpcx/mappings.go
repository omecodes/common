package grpcx

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"net/http"
)

type endpointMapping struct {
	Name   string
	Mapper Mapper
}

type Mapper func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error

type MuxWrapper func(mux *runtime.ServeMux) http.Handler

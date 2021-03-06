package grpcx

import (
	"context"
	"github.com/omecodes/common/netx"
	"google.golang.org/grpc"
	"net/http"
)

type options struct {
	listenOptions   []netx.ListenOption
	gRPCPort        int
	httpPort        int
	grpcSession     bool
	grpcOpts        []grpc.ServerOption
	endpointMappers map[string]endpointMapping
	middlewareList  []func(handler http.Handler) http.Handler
	authFunc        func(ctx context.Context) (context.Context, error)
}

type Option func(opts *options)

func ListenOptions(no ...netx.ListenOption) Option {
	return func(opts *options) {
		opts.listenOptions = append(opts.listenOptions, no...)
	}
}

func Grpc(port int) Option {
	return func(opts *options) {
		opts.gRPCPort = port
	}
}

func Http(port int) Option {
	return func(opts *options) {
		opts.httpPort = port
	}
}

func EndpointMap(name string, mapper Mapper) Option {
	return func(opts *options) {
		if opts.endpointMappers == nil {
			opts.endpointMappers = map[string]endpointMapping{}
		}
		opts.endpointMappers[name] = endpointMapping{
			Name:   name,
			Mapper: mapper,
		}
	}
}

func Middleware(middleware func(handler http.Handler) http.Handler) Option {
	return func(opts *options) {
		opts.middlewareList = append(opts.middlewareList, middleware)
	}
}

func Authentication(authFunc func(ctx context.Context) (context.Context, error)) Option {
	return func(opts *options) {
		opts.authFunc = authFunc
	}
}

func GRPCSession(enable bool) Option {
	return func(opts *options) {
		opts.grpcSession = true
	}
}

func GrpcOptions(gopts ...grpc.ServerOption) Option {
	return func(opts *options) {
		opts.grpcOpts = append(opts.grpcOpts, gopts...)
	}
}

package grpc_gateway

import (
	"github.com/omecodes/common/netx"
	"google.golang.org/grpc"
)

type options struct {
	listenOptions   []netx.ListenOption
	gRPCPort        int
	httpPort        int
	grpcOpts        []grpc.ServerOption
	endpointMappers map[string]endpointMapping
	muxWrappers     []MuxWrapper
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

func MuxWrappers(wrappers ...MuxWrapper) Option {
	return func(opts *options) {
		opts.muxWrappers = wrappers
	}
}

func GrpcOptions(gopts ...grpc.ServerOption) Option {
	return func(opts *options) {
		opts.grpcOpts = append(opts.grpcOpts, gopts...)
	}
}

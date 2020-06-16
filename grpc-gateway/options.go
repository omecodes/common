package grpc_gateway

import (
	"github.com/omecodes/common/netx"
	"google.golang.org/grpc"
)

type options struct {
	listenOptions []netx.ListenOption
	GRPC          int
	*gatewayConfigs
	grpcOpts []grpc.ServerOption
}

type Option func(opts *options)

func ListenOptions(no ...netx.ListenOption) Option {
	return func(opts *options) {
		opts.listenOptions = append(opts.listenOptions, no...)
	}
}

func Grpc(port int) Option {
	return func(opts *options) {
		opts.GRPC = port
	}
}

func Http(port int) Option {
	return func(opts *options) {
		opts.port = port
	}
}

func Gateway(mapper Mapper, wrappers ...MuxWrapper) Option {
	return func(opts *options) {
		if opts.gatewayConfigs == nil {
			opts.gatewayConfigs = new(gatewayConfigs)
		}
		opts.gatewayConfigs.muxWrappers = wrappers
		opts.gatewayConfigs.mapper = mapper
	}
}

func GrpcOptions(gopts ...grpc.ServerOption) Option {
	return func(opts *options) {
		opts.grpcOpts = append(opts.grpcOpts, gopts...)
	}
}

package auth

import (
	"context"
	"fmt"
	"github.com/zoenion/common/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"path"
	"strings"
	"time"
)

//
type gRPCServerAccessAuthentication struct {
	access  string
	methods []string
}

func (gi *gRPCServerAccessAuthentication) InterceptUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	methodName := path.Base(info.FullMethod)
	var (
		rsp interface{}
		err error
	)

	for _, method := range gi.methods {
		if method == methodName {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				err = errors.Forbidden
			}

			authorizationValues := md.Get("authorization")
			if len(authorizationValues) == 0 {
				err = errors.Forbidden
			}

			authorization := strings.TrimPrefix(authorizationValues[0], "Access ")
			if authorization != gi.access {
				err = errors.Forbidden
			}
			break
		}
	}

	if err == nil {
		rsp, err = handler(ctx, req)
	}

	log.Printf("gRPC request - Method:%s\tDuration:%s\tError:%v\n",
		methodName,
		time.Since(start),
		err)

	return rsp, err
}

func (gi *gRPCServerAccessAuthentication) InterceptStream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	start := time.Now()
	methodName := path.Base(info.FullMethod)

	var err error
	for _, method := range gi.methods {
		if method == methodName {
			md, ok := metadata.FromIncomingContext(ss.Context())
			if !ok {
				err = errors.Forbidden
			}

			authorizationValues := md.Get("authorization")
			if len(authorizationValues) == 0 {
				err = errors.Forbidden
			}

			authorization := strings.TrimPrefix(authorizationValues[0], "Access ")
			if authorization != gi.access {
				err = errors.Forbidden
			}
			break
		}
	}
	if err == nil {
		err = handler(srv, newWrappedStream(ss))
	}

	log.Printf("gRPC request - Method:%s\tDuration:%s\tError:%v\n", methodName, time.Since(start), err)
	return err
}

func NewGRPCServerAccessAuthentication(accessKey, accessSecret string, methods ...string) *gRPCServerAccessAuthentication {
	return &gRPCServerAccessAuthentication{
		access:  fmt.Sprintf("%s:%s", accessKey, accessSecret),
		methods: methods,
	}
}

type gRPCServerAuthenticationInterceptor struct {
	methodAuthenticator GRPCMethodAuthenticator
}

func (gi *gRPCServerAuthenticationInterceptor) InterceptUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	methodName := path.Base(info.FullMethod)
	var (
		rsp interface{}
		err error
	)

	err = gi.methodAuthenticator(ctx, methodName)
	if err == nil {
		rsp, err = handler(ctx, req)
	}

	log.Printf("gRPC request - Method:%s\tDuration:%s\tError:%v\n",
		methodName,
		time.Since(start),
		err)

	return rsp, err
}

func (gi *gRPCServerAuthenticationInterceptor) InterceptStream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	start := time.Now()
	methodName := path.Base(info.FullMethod)

	err := gi.methodAuthenticator(ss.Context(), methodName)
	if err == nil {
		err = handler(srv, newWrappedStream(ss))
	}

	log.Printf("gRPC request - Method:%s\tDuration:%s\tError:%v\n",
		methodName,
		time.Since(start),
		err)
	return err
}

func NewGRPCServerAuthenticationInterceptor(ma GRPCMethodAuthenticator) *gRPCServerAuthenticationInterceptor {
	return &gRPCServerAuthenticationInterceptor{
		methodAuthenticator: ma,
	}
}

// logger is to mock a sophisticated logging system. To simplify the example, we just print out the content.
func logger(format string, a ...interface{}) {
	fmt.Printf("LOG:\t"+format+"\n", a...)
}

// wrappedStream wraps around the embedded grpc.ServerStream, and intercepts the RecvMsg and
// SendMsg method call.
type wrappedStream struct {
	grpc.ServerStream
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	logger("Receive a message (Type: %T) at %s", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	logger("Send a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.SendMsg(m)
}

func newWrappedStream(s grpc.ServerStream) grpc.ServerStream {
	return &wrappedStream{s}
}

type GRPCMethodAuthenticator func(ctx context.Context, method string) error

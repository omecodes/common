package auth

import (
	"context"
	authpb "github.com/zoenion/common/proto/auth"
	"google.golang.org/grpc"
)

type gRPCClientBasicAuthentication struct {
	credentials *authpb.Credentials
}

func (g *gRPCClientBasicAuthentication) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Basic ",
	}, nil
}

func (g *gRPCClientBasicAuthentication) RequireTransportSecurity() bool {
	return true
}

func NewGRPCBasicAuthentication(user, password string) *gRPCClientBasicAuthentication {
	return &gRPCClientBasicAuthentication{
		credentials: &authpb.Credentials{
			Username: user,
			Password: password,
		},
	}
}

type gRPCClientTokenAuth struct {
	token string
}

func (g *gRPCClientTokenAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + g.token,
	}, nil
}

func (g *gRPCClientTokenAuth) RequireTransportSecurity() bool {
	return true
}

func NewGRPCTokenAuthentication(t string) *gRPCClientTokenAuth {
	return &gRPCClientTokenAuth{token: t}
}

type gRPCClientChallengeAuthentication struct {
	credentials *authpb.Credentials
	nonce       []byte
}

func (g *gRPCClientChallengeAuthentication) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Zoenion ",
	}, nil
}

func (g *gRPCClientChallengeAuthentication) RequireTransportSecurity() bool {
	return false
}

func NewGRPCChallengeAuthentication(credentials *authpb.Credentials, nonce []byte) *gRPCClientChallengeAuthentication {
	return &gRPCClientChallengeAuthentication{
		credentials: credentials,
		nonce:       nonce,
	}
}

// GRPCDialOptions
func GRPCDialOptions(ctx context.Context) ([]grpc.DialOption, error) {
	return nil, nil
}

// GRPCListenOptions
func GRPCListenOptions(ctx context.Context) ([]grpc.DialOption, error) {
	return nil, nil
}

type gRPCServerAuthenticationInterceptor struct {
	credentials *authpb.Credentials
}

func (gi *gRPCServerAuthenticationInterceptor) gRPCInterceptUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	return nil, nil
}

func (gi *gRPCServerAuthenticationInterceptor) gRPCInterceptStream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return nil
}

func NewGRPCServerAuthenticationInterceptor(credentials *authpb.Credentials) *gRPCServerAuthenticationInterceptor {
	return &gRPCServerAuthenticationInterceptor{
		credentials: credentials,
	}
}

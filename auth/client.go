package auth

import (
	"context"
	"fmt"
)

type gRPCClientApiAccessCredentials struct {
	key, secret string
}

func (g *gRPCClientApiAccessCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": fmt.Sprintf("Access %s:%s", g.key, g.secret),
	}, nil
}

func (g *gRPCClientApiAccessCredentials) RequireTransportSecurity() bool {
	return true
}

func NewGRPCClientApiAccessCredentials(key, secret string) *gRPCClientApiAccessCredentials {
	return &gRPCClientApiAccessCredentials{
		key:    key,
		secret: secret,
	}
}

type gRPCClientBasicAuthentication struct {
	user, password string
}

func (g *gRPCClientBasicAuthentication) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": fmt.Sprintf("Basic %s:%s", g.user, g.password),
	}, nil
}

func (g *gRPCClientBasicAuthentication) RequireTransportSecurity() bool {
	return true
}

func NewGRPCBasicAuthentication(user, password string) *gRPCClientBasicAuthentication {
	return &gRPCClientBasicAuthentication{
		user:     user,
		password: password,
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

func NewGRPCClientTokenAuthentication(t string) *gRPCClientTokenAuth {
	return &gRPCClientTokenAuth{token: t}
}

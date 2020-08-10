package grpcx

import (
	"context"
)

type ctxProxyCredentials struct{}
type ctxCredentials struct{}
type ctxOauth2Token struct{}
type ctxJwt struct{}

func ContextWithJWT(ctx context.Context, j string) context.Context {
	return context.WithValue(ctx, ctxJwt{}, j)
}

func ContextWithCredentials(ctx context.Context, c *Credentials) context.Context {
	return context.WithValue(ctx, ctxCredentials{}, c)
}

func ContextWithProxyCredentials(ctx context.Context, credentials2 *ProxyCredentials) context.Context {
	return context.WithValue(ctx, ctxProxyCredentials{}, credentials2)
}

func ContextWithOauth2Token(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, ctxOauth2Token{}, token)
}

func CredentialsFromContext(ctx context.Context) *Credentials {
	o := ctx.Value(ctxCredentials{})
	if o == nil {
		return nil
	}
	return o.(*Credentials)
}

func ProxyCredentialsFromContext(ctx context.Context) *ProxyCredentials {
	o := ctx.Value(ctxProxyCredentials{})
	if o == nil {
		return nil
	}
	return o.(*ProxyCredentials)
}

func JWTFromContext(ctx context.Context) string {
	o := ctx.Value(ctxJwt{})
	if o == nil {
		return ""
	}

	return o.(string)
}

func GetOauth2Token(ctx context.Context) string {
	o := ctx.Value(ctxOauth2Token{})
	if o == nil {
		return ""
	}

	return o.(string)
}

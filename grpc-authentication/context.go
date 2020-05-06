package ga

import (
	"context"
)

type proxyCredentials struct{}
type credentials struct{}
type jwt struct{}

func ContextWithJWT(ctx context.Context, j string) context.Context {
	return context.WithValue(ctx, jwt{}, j)
}

func ContextWithCredentials(ctx context.Context, c *Credentials) context.Context {
	return context.WithValue(ctx, credentials{}, c)
}

func ContextWithProxyCredentials(ctx context.Context, credentials2 *ProxyCredentials) context.Context {
	return context.WithValue(ctx, proxyCredentials{}, credentials2)
}

func CredentialsFromContext(ctx context.Context) *Credentials {
	o := ctx.Value(credentials{})
	if o == nil {
		return nil
	}
	return o.(*Credentials)
}

func ProxyCredentialsFromContext(ctx context.Context) *ProxyCredentials {
	o := ctx.Value(proxyCredentials{})
	if o == nil {
		return nil
	}
	return o.(*ProxyCredentials)
}

func JWTFromContext(ctx context.Context) string {
	o := ctx.Value(jwt{})
	if o == nil {
		return ""
	}

	return o.(string)
}

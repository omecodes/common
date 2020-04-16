package authpb

import "context"

type token struct{}

func TokenFromContext(ctx context.Context) *JWT {
	o := ctx.Value(token{})
	if c, ok := o.(*JWT); ok {
		return c
	}
	return nil
}

func ContextWithToken(ctx context.Context, t *JWT) context.Context {
	return context.WithValue(ctx, token{}, t)
}

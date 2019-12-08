package authpb

import "context"

type ctxType string

const (
	jwt = "auth-jwt"
)

func ContextWithJWT(ctx context.Context, t *JWT) context.Context {
	return context.WithValue(ctx, jwt, t)
}

func GetContextJWT(ctx context.Context) *JWT {
	o := ctx.Value(jwt)
	if o == nil {
		return nil
	}

	j, castOK := o.(*JWT)
	if !castOK {
		return nil
	}

	return j
}

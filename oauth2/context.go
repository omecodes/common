package oauth2

import "context"

type token struct{}

func ContextWithToken(parent context.Context, t *Token) context.Context {
	return context.WithValue(parent, token{}, t)
}

func TokenFromContext(ctx context.Context) *Token {
	o := ctx.Value(token{})
	if o == nil {
		return nil
	}
	return o.(*Token)
}

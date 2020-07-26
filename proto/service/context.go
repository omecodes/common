package pb

import "context"

type ctxConnectionPool struct{}

func ContextWithConnectionPool(parent context.Context, connectionPool ConnectionPool) context.Context {
	return context.WithValue(parent, ctxConnectionPool{}, connectionPool)
}

func GetConnectionPool(ctx context.Context) ConnectionPool {
	o := ctx.Value(ctxConnectionPool{})
	if o == nil {
		return nil
	}
	return o.(ConnectionPool)
}

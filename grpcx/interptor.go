package grpcx

import "context"

type Interceptor interface {
	Intercept(ctx context.Context) (context.Context, error)
}

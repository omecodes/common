package app

import "context"

type contextKey string

const (
	ctxApp contextKey = "app"
)

func WithContext(parent context.Context, a *App) context.Context {
	return context.WithValue(parent, ctxApp, a)
}

func GetApp(ctx context.Context) *App {
	val := ctx.Value(ctxApp)
	if val == nil {
		return nil
	}

	a, ok := val.(*App)
	if !ok {
		return nil
	}

	return a
}

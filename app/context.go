package app

import (
	"context"
	"github.com/zoenion/common/conf"
)

type contextKey string

const (
	ctxApp contextKey = "app"
)

func ContextWithApp(parent context.Context, a *App) context.Context {
	return context.WithValue(parent, ctxApp, a)
}

func FromContext(ctx context.Context) *App {
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

func ConfigFromContext(ctx context.Context, item ConfigType) conf.Map {
	app := FromContext(ctx)
	cfg := app.configs.GetConf(item.String())
	return cfg
}

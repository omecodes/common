package app

import (
	"context"
	"github.com/omecodes/common/jcon"
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

func ConfigFromContext(ctx context.Context, item ConfigType) jcon.Map {
	app := FromContext(ctx)
	cfg := app.configs.GetConf(item.String())
	return cfg
}

func Oauth2ProviderConfig(ctx context.Context, providerName string) jcon.Map {
	app := FromContext(ctx)
	if app == nil {
		return nil
	}

	cfg := app.configs.GetConf(ConfigOauth2Providers.String())
	if cfg == nil {
		return nil
	}

	return cfg.GetConf(providerName)
}

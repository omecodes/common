package service

import (
	"context"
	"crypto"
	"crypto/x509"
	"github.com/zoenion/common/service/pb"
)

type contextKey string

const (
	ctxBox = contextKey("box")
)

func CACertificate(ctx context.Context) *x509.Certificate {
	val := ctx.Value(ctxBox)
	if val == nil {
		return nil
	}
	return val.(*Box).caCert
}

func Certificate(ctx context.Context) *x509.Certificate {
	val := ctx.Value(ctxBox)
	if val == nil {
		return nil
	}
	return val.(*Box).cert
}

func PrivateKey(ctx context.Context) crypto.PrivateKey {
	val := ctx.Value(ctxBox)
	if val == nil {
		return nil
	}
	return val.(*Box).privateKey
}

func Registry(ctx context.Context) pb.Registry {
	val := ctx.Value(ctxBox)
	if val == nil {
		return nil
	}
	return val.(*Box).registry
}

func Namespace(ctx context.Context) string {
	val := ctx.Value(ctxBox)
	if val == nil {
		return ""
	}
	return val.(*Box).params.Namespace
}

func Name(ctx context.Context) string {
	val := ctx.Value(ctxBox)
	if val == nil {
		return ""
	}
	return val.(*Box).params.Name
}

func Dir(ctx context.Context) string {
	val := ctx.Value(ctxBox)
	if val == nil {
		return ""
	}
	return val.(*Box).params.Dir
}

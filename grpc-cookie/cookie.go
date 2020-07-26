package grpc_cookie

import (
	"context"
	"github.com/omecodes/common/errors"
	"google.golang.org/grpc/metadata"
	"net/http"
)

func Get(ctx context.Context, name string) (*http.Cookie, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.NotFound
	}

	setCookieValues := md.Get("grpcgateway-cookie")
	if len(setCookieValues) == 0 {
		return nil, errors.NotFound
	}

	hr := &http.Request{
		Header: http.Header{},
	}

	hr.Header.Add("Cookie", setCookieValues[0])
	return hr.Cookie(name)
}

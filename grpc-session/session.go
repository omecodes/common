package gs

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/zoenion/common/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net/http"
)

type ctxSession struct{}

const (
	gRPCSetCookie         = "set-cookie"
	gRPCHeaderSetCookie   = "Grpc-Metadata-Set-Cookie"
	gRPCHeaderContentType = "Grpc-Metadata-Content-Type"
	gRPCHeaderCookie      = "grpcgateway-cookie"

	httpHeaderSetCookie = "Set-Cookie"
	httpHeaderCookie    = "Cookie"
)

type Session struct {
	Values map[interface{}]interface{}
	Name   string
	Store  *sessions.CookieStore
}

func (s *Session) Save(ctx context.Context) error {
	encoded, err := securecookie.EncodeMulti(s.Name, s.Values, s.Store.Codecs...)
	if err != nil {
		return err
	}
	value := sessions.NewCookie(s.Name, encoded, s.Store.Options).String()
	return grpc.SendHeader(ctx, metadata.Pairs(gRPCSetCookie, value))
}

func SessionFromContext(ctx context.Context, store *sessions.CookieStore, name string) (*Session, error) {
	o := ctx.Value(ctxSession{})
	if o != nil {
		s, ok := o.(*Session)
		if !ok {
			return nil, errors.New("invalid type for context 'session' key")
		}
		return s, nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.NotFound
	}

	setCookieValues := md.Get(gRPCHeaderCookie)
	if len(setCookieValues) == 0 {
		return &Session{
			Values: map[interface{}]interface{}{},
			Name:   name,
			Store:  store,
		}, nil
	}
	hr := &http.Request{
		Header: http.Header{},
	}

	hr.Header.Add(httpHeaderCookie, setCookieValues[0])
	c, err := hr.Cookie(name)
	if err != nil {
		return nil, err
	}

	values := make(map[interface{}]interface{})

	err = securecookie.DecodeMulti(name, c.Value, &values, store.Codecs...)
	if err == nil {
		s := &Session{
			Values: values,
			Name:   name,
			Store:  store,
		}
		ctx = context.WithValue(ctx, ctxSession{}, s)
		return s, nil
	}
	return nil, err
}

func SetCookieFromGRPCMetadata(ctx context.Context, w http.ResponseWriter, msg proto.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if ok {
		values := md.HeaderMD.Get(gRPCSetCookie)
		if len(values) > 0 {
			w.Header().Set(httpHeaderSetCookie, values[0])
		}
	}
	w.Header().Del(gRPCHeaderSetCookie)
	w.Header().Del(gRPCHeaderContentType)
	return nil
}

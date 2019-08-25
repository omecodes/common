package xhttp

import (
	"context"
	"github.com/gorilla/sessions"
	"net/http"
)

type Cookie struct {
	http.Cookie
}

type Session interface {
	Set(name string, value interface{})
	Get(name string) interface{}
	Delete(name string)
	Save() error
}

func SetCookie(ctx context.Context, cookie *Cookie) {
	if cookie == nil {
		return
	}
	rw := ctx.Value(CtxResponseWriter)
	http.SetCookie(rw.(http.ResponseWriter), &cookie.Cookie)
}

func GetSession(ctx context.Context, r *http.Request, name string) Session {
	rw := ctx.Value(CtxResponseWriter)
	cs := ctx.Value(CtxCCookiesStore)
	return newSession(name, r, rw.(http.ResponseWriter), cs.(*sessions.CookieStore))
}

type session struct {
	store       *sessions.CookieStore
	httpSession *sessions.Session
	r           *http.Request
	w           http.ResponseWriter
}

func newSession(name string, r *http.Request, w http.ResponseWriter, store *sessions.CookieStore) Session {
	s := new(session)
	s.store = store
	s.r = r
	s.w = w
	s.httpSession, _ = store.Get(r, name)
	return s
}

func (s *session) Set(name string, value interface{}) {
	s.httpSession.Values[name] = value
}

func (s *session) Get(name string) interface{} {
	v, ok := s.httpSession.Values[name]
	if !ok {
		return nil
	}
	return v
}

func (s *session) Delete(key string) {
	delete(s.httpSession.Values, key)
}

func (s *session) Save() error {
	return s.httpSession.Save(s.r, s.w)
}

package oauth2

import (
	"encoding/gob"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/omecodes/common/httpx"
	"github.com/omecodes/common/utils/log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type ConfigProvider func() (*Config, error)

func init() {
	gob.Register(Token{})
}

var states = &sync.Map{}

type AuthorizedHandleFunc func(t *Token, continueURL string, w http.ResponseWriter, r *http.Request)

type AuthenticationRequiredFunc func(r *http.Request) bool

type workflow struct {
	callbackURL        string
	callbackRequestURI string
	configProvider     ConfigProvider
	continueURL        string
	authRequiredFunc   AuthenticationRequiredFunc
	handlerFunc        AuthorizedHandleFunc
}

func (m *workflow) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == m.callbackURL {
			m.authorized(w, r)
			return
		}

		if m.authRequiredFunc(r) {
			m.login(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *workflow) authorized(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	authError := q.Get(ParamError)
	if authError != "" {
		authErrorDesc := q.Get(ParamErrorDescription)
		log.Info("failed to authenticate user with Ome", log.Field("error", authError), log.Field("description", authErrorDesc))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	state := q.Get(ParamState)
	code := q.Get(ParamCode)

	if state == "" || code == "" {
		log.Info("OAuth incomplete Ome response: expected code and state")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	o, ok := states.Load(state)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	oc := o.(*Config)
	states.Delete(state)

	client := NewClient(oc)

	token, err := client.GetAccessToken(code)
	if err != nil {
		log.Error("failed to get JWT from Ome server", log.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	m.handlerFunc(token, m.continueURL, w, r)
}

func (m *workflow) login(w http.ResponseWriter, r *http.Request) {
	config, err := m.configProvider()
	if err != nil {
		log.Error("could not get", log.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	config.CallbackURL = m.callbackURL

	q := r.URL.Query()
	m.continueURL = q.Get("continue")

	states.Store(config.State, config)

	client := NewClient(config)
	authorizeURL, err := client.GetURLAuthorizationURL()
	if err != nil {
		log.Error("failed to construct OAuth authorize URL", log.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httpx.Redirect(w, &httpx.RedirectURL{
		URL:         authorizeURL,
		Code:        http.StatusUnauthorized,
		ContentType: "text/html",
	})
}

func Workflow(callbackURI string, configProvider ConfigProvider, authRequiredFunc AuthenticationRequiredFunc, handlerFunc AuthorizedHandleFunc) (mux.MiddlewareFunc, error) {
	u, err := url.Parse(callbackURI)
	if err != nil {
		return nil, err
	}

	m := &workflow{
		callbackURL:        callbackURI,
		callbackRequestURI: u.RequestURI(),
		configProvider:     configProvider,
		authRequiredFunc:   authRequiredFunc,
		handlerFunc:        handlerFunc,
	}
	return m.middleware, nil
}

type bearerDecoder struct {
	codecs []securecookie.Codec
}

func (bc bearerDecoder) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization != "" {
			accessToken := strings.TrimLeft(authorization, "Bearer ")
			var token Token
			err := securecookie.DecodeMulti("", accessToken, &token, bc.codecs...)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			ctx := r.Context()
			ctx = ContextWithToken(ctx, &token)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func BearerHeaderDecoder(codecs ...securecookie.Codec) *bearerDecoder {
	return &bearerDecoder{codecs: codecs}
}

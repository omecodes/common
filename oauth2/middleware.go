package oauth2

import (
	"encoding/gob"
	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
	"github.com/omecodes/common/httpx"
	"github.com/omecodes/common/log"
	"github.com/zoenion/common/oauth2"
	"net/http"
	"net/url"
	"sync"
	"time"
)

func init() {
	gob.Register(Token{})
}

var states = &sync.Map{}

type AuthorizedHandleFunc func(state string, t *Token, w http.ResponseWriter, r *http.Request)

type middleware struct {
	config             *Config
	triggerEndpoint    string
	authorizedEndpoint string
	codecs             []securecookie.Codec
	cookieName         string
	continueURL        string
}

func (m *middleware) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == m.triggerEndpoint {
			m.login(w, r)
			return
		}

		if r.URL.Path == m.authorizedEndpoint {
			m.authorized(w, r)
			return
		}
		cookie, _ := r.Cookie(m.cookieName)
		if cookie != nil && cookie.Value != "" {
			token := &Token{}
			err := securecookie.DecodeMulti("", cookie.Value, token, m.codecs...)
			if err != nil {
				log.Error("could not decode cookie", err)

			} else {
				ctx := ContextWithToken(r.Context(), token)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (m *middleware) authorized(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	authError := q.Get(ParamError)
	if authError != "" {
		authErrorDesc := q.Get(oauth2.ParamErrorDescription)
		log.Info("failed to authenticate user with Ome", log.Field("error", authError), log.Field("description", authErrorDesc))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	state := q.Get(oauth2.ParamState)
	code := q.Get(oauth2.ParamCode)

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
		log.Error("failed to get JWT from Ome server", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookieValue, err := securecookie.EncodeMulti("", token, m.codecs...)
	if err != nil {
		log.Error("could not encode token to cookie", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     m.cookieName,
		Value:    cookieValue,
		Expires:  time.Now().Add(time.Hour * 30),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	httpx.Redirect(w, &httpx.RedirectURL{
		URL:         m.continueURL,
		Code:        http.StatusOK,
		ContentType: "text/html",
	})
}

func (m *middleware) login(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	m.continueURL = q.Get("continue")

	state := uuid.New().String()
	m.config.State = state

	states.Store(state, m.config)

	client := NewClient(m.config)
	authorizeURL, err := client.GetURLAuthorizationURL()
	if err != nil {
		log.Error("failed to construct OAuth authorize URL", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httpx.Redirect(w, &httpx.RedirectURL{
		URL:         authorizeURL,
		Code:        http.StatusUnauthorized,
		ContentType: "text/html",
	})
}

func WebAuthenticator(config *Config, triggerEndpoint string, cookieName string, codecs ...securecookie.Codec) (func(http.Handler) http.Handler, error) {
	m := &middleware{
		config:          config,
		triggerEndpoint: triggerEndpoint,
		cookieName:      cookieName,
		codecs:          codecs,
	}
	u, err := url.Parse(config.CallbackURL)
	if err != nil {
		return nil, err
	}
	m.authorizedEndpoint = u.Path
	return m.middleware, nil
}

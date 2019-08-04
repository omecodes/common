package http_helper

import (
	"fmt"
	auth "github.com/abbot/go-http-auth"
	"net/http"
)

type AuthProviderFunc func(string) string

type Authentication struct {
	digest *auth.DigestAuth
	basic  *auth.BasicAuth

	provider     AuthProviderFunc
	name         string
	providerFunc AuthProviderFunc
}

func (ba *Authentication) Wrap(next http.HandlerFunc) http.HandlerFunc {
	if ba.digest != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			if username, _ := ba.digest.CheckAuth(r); username == "" {
				ba.digest.RequireAuth(w, r)
			} else {
				r.Header.Set("X-Authenticated-Username", username)
				next(w, r)
			}
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if user := ba.basic.CheckAuth(r); user == "" {
			WriteResponse(w, 401, &RequireAuth{
				Realm: ba.name,
				Type:  "Basic",
			})
		} else {
			r.Header.Set("X-Authenticated-Username", user)
		}
		next.ServeHTTP(w, r)
	}
}

func NewAuthentication(name string, digest bool, providerFunc AuthProviderFunc) *Authentication {
	a := &Authentication{
		provider:     providerFunc,
		name:         name,
		providerFunc: providerFunc,
	}

	if digest {
		a.digest = auth.NewDigestAuthenticator(name, func(username, realm string) string {
			return providerFunc(fmt.Sprintf("digest://%s.%s", username, realm))
		})
	} else {
		a.basic = &auth.BasicAuth{
			Realm: name,
			Secrets: func(username, real string) string {
				return providerFunc(username)
			},
		}
	}
	return a
}

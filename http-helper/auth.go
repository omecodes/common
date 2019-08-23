package http_helper

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	auth "github.com/abbot/go-http-auth"
	"github.com/zoenion/common/errors"
	authpb "github.com/zoenion/common/proto/auth"
	"log"
	"net/http"
	"strings"
)

// CredentialsProvider
type CredentialsProvider func(args ...string) *authpb.Credentials

type JwtVerifier func(string) bool

// DigestAuthenticationMiddleware
type DigestAuthenticationMiddleware struct {
	digest   *auth.DigestAuth
	realm    string
	wrappers []RequestWrapper
}

func (dam *DigestAuthenticationMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if username, _ := dam.digest.CheckAuth(r); username == "" {
			dam.digest.RequireAuth(w, r)
		} else {
			for _, wrapper := range dam.wrappers {
				r = wrapper(r)
			}
			next(w, r)
		}
	}
}

func NewDigestAuthenticationMiddleware(realm string, provider CredentialsProvider, wrappers ...RequestWrapper) *DigestAuthenticationMiddleware {
	a := &DigestAuthenticationMiddleware{
		realm:    realm,
		wrappers: wrappers,
	}
	a.digest = auth.NewDigestAuthenticator(realm, func(username, realm string) string {
		credentials := provider(username)
		if credentials == nil {
			return ""
		}

		ha1Str := fmt.Sprintf("%s:%s:%s", username, realm, credentials.Password)
		h := md5.New()
		h.Write([]byte(ha1Str))
		ha1Bytes := h.Sum(nil)
		return hex.EncodeToString(ha1Bytes)
	})
	return a
}

// BasicAuthenticationMiddleware
type BasicAuthenticationMiddleware struct {
	provider CredentialsProvider
	wrappers []RequestWrapper
	realm    string
}

func (bam *BasicAuthenticationMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	reqAuth := &RequireAuth{
		Realm: bam.realm,
		Type:  "Basic",
	}
	return func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok {
			WriteResponse(w, 401, reqAuth)
			return
		}

		c := bam.provider(user)
		if c == nil || (c.Username != user && c.Email != user) || c.Password != password {
			WriteResponse(w, 401, reqAuth)
			return
		}

		for _, wrapper := range bam.wrappers {
			r = wrapper(r)
		}
		next.ServeHTTP(w, r)
	}
}

func NewBasicAuthenticationMiddleware(realm string, provider CredentialsProvider, wrappers ...RequestWrapper) *BasicAuthenticationMiddleware {
	return &BasicAuthenticationMiddleware{
		realm:    realm,
		provider: provider,
		wrappers: wrappers,
	}
}

// BearerAuthenticationMiddleware
type BearerAuthenticationMiddleware struct {
	jwtVerifier JwtVerifier
	wrappers    []RequestWrapper
}

func (bam *BearerAuthenticationMiddleware) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if bam.jwtVerifier(authorization) {
			for _, wrapper := range bam.wrappers {
				r = wrapper(r)
			}
			next(w, r)
			return
		}
		WriteError(w, errors.HttpForbidden)
	}
}

func NewBearerAuthenticationMiddleware(jwtVerifier JwtVerifier, wrappers ...RequestWrapper) *BearerAuthenticationMiddleware {
	return &BearerAuthenticationMiddleware{
		jwtVerifier: jwtVerifier,
		wrappers:    wrappers,
	}
}

// ProxyAuthentication
type APIAccessAuthorization struct {
	realm       string
	key, secret string
	wrappers    []RequestWrapper
}

func (pam *APIAccessAuthorization) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authentication := r.Header.Get("X-Api-Key")
		if authentication == "" {
			log.Println("Api access refused")
			WriteResponse(w, http.StatusUnauthorized, "API access refused")
			return
		}

		parts := strings.Split(authentication, ":")
		if len(parts) != 2 || pam.key != parts[0] || pam.secret != parts[1] {
			log.Println("Api access refused")
			WriteResponse(w, http.StatusUnauthorized, "API access refused")
			return
		}

		for _, wrapper := range pam.wrappers {
			r = wrapper(r)
		}
		next(w, r)
	}
}

func NewAPIAuthenticationMiddleware(realm string, key string, secret string, wrappers ...RequestWrapper) *APIAccessAuthorization {
	return &APIAccessAuthorization{
		realm:    realm,
		key:      key,
		secret:   secret,
		wrappers: wrappers,
	}
}

// Gateway wraps autho
type GRPCTranslatorAuthorization struct {
	secret string
}

func (gt *GRPCTranslatorAuthorization) Handle(new http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("authorization")
		r.Header.Set("authorization", fmt.Sprintf("Gateway %s:%s", gt.secret, authorization))
	}
}

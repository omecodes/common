package interceptors

import (
	"fmt"
	"net/http"
)

// GRPCTranslatorAuthorization
type GRPCTranslatorAuthorization struct {
	secret string
}

func (gt *GRPCTranslatorAuthorization) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("authorization")
		/*if authorization == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}*/
		r.Header.Set("authorization", fmt.Sprintf("Gateway %s/%s", gt.secret, authorization))
		next.ServeHTTP(w, r)
	}
}

func NewGRPCTranslatorAuthorization(secret string) *GRPCTranslatorAuthorization {
	return &GRPCTranslatorAuthorization{
		secret: secret,
	}
}

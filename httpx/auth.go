package httpx

import (
	"encoding/base64"
	ga "github.com/omecodes/common/grpc-authentication"
	"net/http"
	"strings"
)

func ProxyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Proxy-Authorization")
		if authorization != "" {
			authorization = strings.TrimPrefix(authorization, "Basic ")

			decodedBytes, err := base64.StdEncoding.DecodeString(authorization)
			if err != nil {
				w.WriteHeader(http.StatusProxyAuthRequired)
				return
			}

			var key string
			var secret string

			splits := strings.Split(string(decodedBytes), ":")
			key = splits[0]
			if len(splits) > 1 {
				secret = splits[1]
			}

			ctx := r.Context()
			ctx = ga.ContextWithProxyCredentials(ctx, &ga.ProxyCredentials{
				Key:    key,
				Secret: secret,
			})
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func BasicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, _ := r.BasicAuth()
		if user != "" || password != "" {
			ctx := r.Context()
			ctx = ga.ContextWithCredentials(ctx, &ga.Credentials{
				Username: user,
				Password: password,
			})
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

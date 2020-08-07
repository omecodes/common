package authpb

import (
	"github.com/gorilla/securecookie"
	"github.com/omecodes/common/log"
	"net/http"
	"strings"
)

type bearerInterceptor struct {
	codecs   []securecookie.Codec
	verifier TokenVerifier
}

func (atv *bearerInterceptor) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authorizationHeader, "Bearer ") {
			accessToken := strings.TrimLeft(authorizationHeader, "Bearer ")

			strJWT, err := ExtractJwtFromAccessToken("", accessToken, atv.codecs...)
			if err != nil {
				log.Error("could not extract jwt from access token", err)
				next.ServeHTTP(w, r)
				return
			}

			t, err := ParseJWT(strJWT)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			state, err := atv.verifier.Verify(r.Context(), t)
			if err != nil {
				log.Error("could not verify JWT", err, log.Field("jwt", strJWT))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if state != JWTState_VALID {
				log.Info("invalid JWT", log.Field("jwt", strJWT))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// enrich context with
			ctx := r.Context()
			ctx = ContextWithToken(ctx, t)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func BearerTokenInterceptor(verifier TokenVerifier, codecs ...securecookie.Codec) *bearerInterceptor {
	return &bearerInterceptor{
		codecs:   codecs,
		verifier: verifier,
	}
}

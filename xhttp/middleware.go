package xhttp

import (
	"fmt"
	"github.com/zoenion/common/log"
	"net/http"
	"time"
)

func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		log.Info(fmt.Sprintf("[http] %s", r.Method), log.Field("uri", r.RequestURI), log.Field("duration", time.Since(start)))
	})
}

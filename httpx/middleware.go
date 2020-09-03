package httpx

import (
	"context"
	"github.com/omecodes/common/utils/log"
	"net/http"
	"time"
)

type logger struct {
	name   string
	w      http.ResponseWriter
	status int
}

func (l *logger) Header() http.Header {
	return l.w.Header()
}

func (l *logger) Write(bytes []byte) (int, error) {
	return l.w.Write(bytes)
}

func (l *logger) WriteHeader(statusCode int) {
	l.status = statusCode
}

func Logger(name string) *logger {
	return &logger{name: name}
}

func (l *logger) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l.w = w
		start := time.Now()
		next.ServeHTTP(l, r)
		duration := time.Since(start)

		if l.status == http.StatusOK || l.status == 0 {
			log.Info(
				r.Method+" "+r.RequestURI,
				log.Field("params", r.URL.RawQuery),
				log.Field("handler", l.name),
				log.Field("duration", duration.String()),
			)
		} else {
			log.Error(
				r.Method+" "+r.RequestURI,
				log.Field("params", r.URL.RawQuery),
				log.Field("handler", l.name),
				log.Field("duration", duration.String()),
			)
		}
	})
}

type contextUpdater struct {
	updateFunc func(ctx context.Context) context.Context
}

func ContextUpdater(updateFunc func(ctx context.Context) context.Context) *contextUpdater {
	return &contextUpdater{updateFunc: updateFunc}
}

func (updater *contextUpdater) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := updater.updateFunc(r.Context())
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

package httpx

import (
	"context"
	"github.com/omecodes/common/utils/log"
	"net/http"
	"time"
)

type statusCatcher struct {
	status int
	w      http.ResponseWriter
}

func (catcher *statusCatcher) Header() http.Header {
	return catcher.w.Header()
}

func (catcher *statusCatcher) Write(bytes []byte) (int, error) {
	return catcher.w.Write(bytes)
}

func (catcher *statusCatcher) WriteHeader(statusCode int) {
	catcher.status = statusCode
	catcher.w.WriteHeader(statusCode)
}

type logger struct {
	name string
}

func Logger(name string) *logger {
	return &logger{name: name}
}

func (l *logger) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := &statusCatcher{
			status: 0,
			w:      w,
		}

		start := time.Now()
		next.ServeHTTP(c, r)
		duration := time.Since(start)

		if c.status == http.StatusOK || c.status == 0 {
			log.Info(
				r.Method+" "+r.RequestURI,
				log.Field("params", r.URL.RawQuery),
				log.Field("handler", l.name),
				log.Field("duration", duration.String()),
			)
		} else {
			log.Error(
				r.Method+" "+r.RequestURI+" "+http.StatusText(c.status),
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

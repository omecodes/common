package xhttp

import (
	"context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

type ContextWrapper func(context.Context) context.Context

type Route struct {
	Name         string
	Method       []string
	Pattern      string
	PathIsPrefix bool
	HandlerFunc  http.HandlerFunc
}

func NewRouter(routes ...Route) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc

		if route.PathIsPrefix {
			sr := router.PathPrefix(route.Pattern).Subrouter()
			sr.Name(route.Name).Methods(route.Method...).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlers.LoggingHandler(os.Stdout, handler).ServeHTTP(w, r)
			}))
		} else {
			router.
				Methods(route.Method...).
				Path(route.Pattern).
				Name(route.Name).
				Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					handlers.LoggingHandler(os.Stdout, handler).ServeHTTP(w, r)
				}))
		}
	}
	return router
}

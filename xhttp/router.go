package xhttp

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
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
			sr.Name(route.Name).Methods(route.Method...).Handler(handler)
		} else {
			router.
				Methods(route.Method...).
				Path(route.Pattern).
				Name(route.Name).
				Handler(handler)
		}
	}
	return router
}

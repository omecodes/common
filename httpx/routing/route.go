package routing

import "net/http"

type Route struct {
	Name         string
	Method       []string
	Pattern      string
	PathIsPrefix bool
	HandlerFunc  http.HandlerFunc
}

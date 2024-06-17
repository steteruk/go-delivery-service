package http

import (
	nethttp "net/http"

	"github.com/gorilla/mux"
)

// Route handles different path routes.
type Route struct {
	Handler func(nethttp.ResponseWriter, *nethttp.Request)
	Method  string
}

// NewRoute creates for handling different path routes.
func NewRoute(routes map[string]Route, router *mux.Router) *mux.Router {
	for url, route := range routes {
		router.HandleFunc(url, route.Handler).Methods(route.Method)
	}

	return router
}

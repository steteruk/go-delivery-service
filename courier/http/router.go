package http

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Router struct {
	routes []Route
}

type Route struct {
	Method  string
	Pattern string
	Handler func(http.ResponseWriter, *http.Request)
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) AddRoute(method, path string, handler func(http.ResponseWriter, *http.Request)) {
	r.routes = append(r.routes, Route{Method: method, Pattern: path, Handler: handler})
}

func (r *Router) Init() *mux.Router {
	newRouter := mux.NewRouter()
	for _, route := range r.routes {
		newRouter.HandleFunc(route.Pattern, route.Handler).Methods(route.Method)
	}

	return newRouter
}

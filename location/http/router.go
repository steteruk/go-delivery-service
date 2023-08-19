package http

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type Router struct {
	url                string
	uuidValidationRule string
	locationHandler    func(http.ResponseWriter, *http.Request)
}

func NewRouter(locationHandler func(http.ResponseWriter, *http.Request)) *Router {
	return &Router{
		url:                "/courier/{courier_id:%s}/location",
		uuidValidationRule: "[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}",
		locationHandler:    locationHandler,
	}
}

func (r *Router) Init() *mux.Router {
	newRouter := mux.NewRouter()
	newRouter.HandleFunc(fmt.Sprintf(r.url, r.uuidValidationRule), r.locationHandler)
	newRouter.Methods("POST")

	return newRouter
}

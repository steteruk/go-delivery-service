package http

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/steteruk/go-delivery-service/location/http/handler"
)

type Router struct {
	url                string
	uuidValidationRule string
}

func NewRouter() *Router {
	return &Router{
		url:                "/courier/{courier_id:%s}/location",
		uuidValidationRule: "[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}",
	}
}

func (r *Router) Init() *mux.Router {
	newRouter := mux.NewRouter()
	fmt.Println(fmt.Sprintf(r.url, r.uuidValidationRule))
	newRouter.HandleFunc(fmt.Sprintf(r.url, r.uuidValidationRule), handler.NewLocationHandler().CourierHandler)
	newRouter.Methods("POST")

	return newRouter
}

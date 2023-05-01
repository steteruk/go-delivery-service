package http

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/steteruk/go-delivery-service/location/http/handler"
	"github.com/steteruk/go-delivery-service/storage/redis"
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
	newRouter.HandleFunc(fmt.Sprintf(r.url, r.uuidValidationRule), handler.NewLocationHandler(redis.NewCourierRepository()).CourierHandler)
	newRouter.Methods("POST")

	return newRouter
}

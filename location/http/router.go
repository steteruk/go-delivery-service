package http

import (
	"github.com/gorilla/mux"
	"github.com/steteruk/go-delivery-service/location/http/handler"
)

func GetRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/courier/{courier_id:[0-9]+}/location", handler.CourierHandler)
	r.Methods("POST")

	return r
}

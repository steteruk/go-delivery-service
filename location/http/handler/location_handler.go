package handler

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Location struct {
	Latitude  float32 `json:"latitude" validate:"required,latitude"`
	Longitude float32 `json:"longitude" validate:"required,longitude"`
}

func CourierHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["courier_id"]
	fmt.Printf(" courier_id = %s\n", id)

	var location Location
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		log.Printf("Request body does not match json format %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := validator.New().Struct(location); err != nil {
		log.Printf("Invalid data for geopoint %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

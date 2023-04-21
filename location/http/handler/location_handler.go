package handler

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"strings"
)

type LocationPayload struct {
	Latitude  float64 `json:"latitude" validate:"required,latitude"`
	Longitude float64 `json:"longitude" validate:"required,longitude"`
}

type ResponseMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type LocationHandler struct {
	validate *validator.Validate
}

func NewLocationHandler() *LocationHandler {
	return &LocationHandler{
		validate: validator.New(),
	}
}

func (lh *LocationHandler) CourierHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := mux.Vars(r)["courier_id"]
	fmt.Printf(" courier_id = %s\n", id)

	locationPayload, response := lh.decodePayload(r.Body)
	if response != nil {
		lh.createErrorResponse(response, w)
		return
	}

	if isValid, response := lh.validatePayload(locationPayload); !isValid {
		lh.createErrorResponse(response, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (lh *LocationHandler) validatePayload(locationPayload *LocationPayload) (isValid bool, response *ResponseMessage) {
	if err := lh.validate.Struct(locationPayload); err != nil {
		errorMessage := ""
		for _, errStruct := range err.(validator.ValidationErrors) {
			message := fmt.Sprintf("Incorrect Value %s %f", errStruct.StructField(), errStruct.Value())
			errorMessage += message + ","
		}
		errorMessage = strings.Trim(errorMessage, ",")
		return false, &ResponseMessage{
			Status:  "Error",
			Message: errorMessage,
		}
	}

	return true, nil
}

func (lh *LocationHandler) decodePayload(payload io.ReadCloser) (locationPayload *LocationPayload, response *ResponseMessage) {
	if err := json.NewDecoder(payload).Decode(&locationPayload); err != nil {

		return nil, &ResponseMessage{
			Status:  "Error",
			Message: fmt.Sprintf("Request body does not match json format %v", err),
		}
	}

	return locationPayload, nil
}

func (lh *LocationHandler) createErrorResponse(response *ResponseMessage, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Failed to encode json response: %v\n", err)
	}
}

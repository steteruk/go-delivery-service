package handler

import (
	"context"
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
	validate          *validator.Validate
	courierRepository CourierRepository
}

func NewLocationHandler(courierRepository CourierRepository) *LocationHandler {
	return &LocationHandler{
		validate:          validator.New(),
		courierRepository: courierRepository,
	}
}

type CourierRepository interface {
	SaveLatestCourierGeoPosition(ctx context.Context, courierID string, latitude, longitude float64) error
}

func (h *LocationHandler) CourierHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	locationPayload, response := h.decodePayload(r.Body)
	if response != nil {
		h.createErrorResponse(response, w)
		return
	}

	if isValid, response := h.validatePayload(locationPayload); !isValid {
		h.createErrorResponse(response, w)
		return
	}

	id := mux.Vars(r)["courier_id"]
	ctx := r.Context()
	err := h.courierRepository.SaveLatestCourierGeoPosition(ctx, id, locationPayload.Latitude, locationPayload.Longitude)
	if err != nil {
		log.Printf("Error saving geodata to storage: %v\n", err)
		h.createErrorResponse(h.getCourierGeoPositionErrorResponse(), w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LocationHandler) validatePayload(locationPayload *LocationPayload) (isValid bool, response *ResponseMessage) {
	if err := h.validate.Struct(locationPayload); err != nil {
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

func (h *LocationHandler) getCourierGeoPositionErrorResponse() (response *ResponseMessage) {
	return &ResponseMessage{
		Status:  "Error",
		Message: fmt.Sprintf("Error saving geodata to storage."),
	}
}

func (h *LocationHandler) decodePayload(payload io.ReadCloser) (locationPayload *LocationPayload, response *ResponseMessage) {
	if err := json.NewDecoder(payload).Decode(&locationPayload); err != nil {

		return nil, &ResponseMessage{
			Status:  "Error",
			Message: fmt.Sprintf("Request body does not match json format %v", err),
		}
	}

	return locationPayload, nil
}

func (h *LocationHandler) createErrorResponse(response *ResponseMessage, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Failed to encode json response: %v\n", err)
	}
}

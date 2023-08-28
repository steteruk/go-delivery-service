package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/steteruk/go-delivery-service/location/domain"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type ResponseMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type LocationHandler struct {
	courierService domain.CourierLocationServiceInterface
	validator      *validator.Validate
}

func NewLocationHandler(courierService domain.CourierLocationServiceInterface) *LocationHandler {
	return &LocationHandler{
		courierService: courierService,
		validator:      validator.New(),
	}
}

type CourierRepository interface {
	SaveLatestCourierGeoPosition(ctx context.Context, courierID string, latitude, longitude float64) error
}

func (h *LocationHandler) CourierHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	courierLocation, response := h.decodePayload(r.Body)
	if response != nil {
		h.createErrorResponse(response, w)
		return
	}

	courierLocation.CreatedAt = time.Now()
	courierLocation.CourierID = mux.Vars(r)["courier_id"]
	if err := h.validateCourierLocation(courierLocation); err != nil {
		log.Printf("Error validate geodata: %v\n", err)
		h.createErrorResponse(prepareErrorResponse(err.Error()), w)
		return
	}

	ctx := r.Context()
	err := h.courierService.SaveLatestCourierLocation(ctx, courierLocation)
	if err != nil {
		log.Printf("Error saving geodata to storage: %v\n", err)
		h.createErrorResponse(prepareErrorResponse("Error saving geodata."), w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LocationHandler) validateCourierLocation(courierLocation *domain.CourierLocation) error {
	err := h.validator.Struct(courierLocation)

	if err == nil {
		return nil
	}
	errorMessage := ""
	for _, errStruct := range err.(validator.ValidationErrors) {
		message := fmt.Sprintf("Incorrect Value %s %f", errStruct.StructField(), errStruct.Value())
		errorMessage += message + ","
	}
	errorMessage = strings.Trim(errorMessage, ",")
	return errors.New(errorMessage)
}

func prepareErrorResponse(massage string) (response *ResponseMessage) {
	return &ResponseMessage{
		Status:  "Error",
		Message: fmt.Sprintf(massage),
	}
}

func (h *LocationHandler) decodePayload(payload io.ReadCloser) (courierLocation *domain.CourierLocation, response *ResponseMessage) {
	if err := json.NewDecoder(payload).Decode(&courierLocation); err != nil {
		return nil, prepareErrorResponse("Request body does not match json format.")
	}

	return courierLocation, nil
}

func (h *LocationHandler) createErrorResponse(response *ResponseMessage, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Failed to encode json response: %v\n", err)
	}
}

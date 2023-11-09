package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/steteruk/go-delivery-service/courier/domain"
	"io"
	"log"
	"net/http"
	"strings"
)

type ResponseMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type SuccessResponseMessage struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Entity  *domain.Courier `json:"entity"`
}

type CourierHandler struct {
	courierService domain.CourierServiceInterface
	validator      *validator.Validate
}

type CourierPayload struct {
	Firstname string `json:"firstname" validate:"required,lte=40"`
}

type GetCourierPayload struct {
	CourierId string `json:"courier_id" validate:"required,uuid"`
}

func NewCourierHandler(courierService domain.CourierServiceInterface) *CourierHandler {
	return &CourierHandler{
		courierService: courierService,
		validator:      validator.New(),
	}
}

func (h *CourierHandler) SaveNewCourierHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	courierPayload, err := h.decodePayload(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.createErrorResponse(err.Error(), w)
		return
	}

	if err = h.validatePayload(courierPayload); err != nil {
		log.Printf("Error validate courier: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		h.createErrorResponse(err.Error(), w)
		return
	}
	courier := &domain.Courier{FirstName: courierPayload.Firstname}
	ctx := r.Context()
	courier, err = h.courierService.SaveNewCourier(ctx, courier)
	if err != nil {
		log.Printf("Error saving courier to storage: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		h.createErrorResponse("Error saving courier.", w)
		return
	}

	w.WriteHeader(http.StatusCreated)
	h.createSuccessResponse("New courier created.", courier, w)
}

func (h *CourierHandler) validatePayload(payload any) error {
	err := h.validator.Struct(payload)

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

func (h *CourierHandler) decodePayload(payload io.ReadCloser) (*CourierPayload, error) {
	var courierPayload CourierPayload
	if err := json.NewDecoder(payload).Decode(&courierPayload); err != nil {
		return nil, fmt.Errorf("row couirier location was not saved: %w", err)
	}

	return &courierPayload, nil
}

func (h *CourierHandler) createErrorResponse(massage string, w http.ResponseWriter) {
	responseMsg := &ResponseMessage{
		Status:  "Error",
		Message: fmt.Sprintf(massage),
	}
	err := json.NewEncoder(w).Encode(responseMsg)
	if err != nil {
		log.Printf("Failed to encode json response: %v\n", err)
	}
}

func (h *CourierHandler) createSuccessResponse(massage string, courier *domain.Courier, w http.ResponseWriter) {
	responseMsg := &SuccessResponseMessage{
		Status:  "Success",
		Message: fmt.Sprintf(massage),
		Entity:  courier,
	}
	err := json.NewEncoder(w).Encode(responseMsg)
	if err != nil {
		log.Printf("Failed to encode json response: %v\n", err)
	}
}

func (h *CourierHandler) GetCourierHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var getCourierPayload GetCourierPayload
	getCourierPayload.CourierId = mux.Vars(r)["courier_id"]
	if err := h.validatePayload(getCourierPayload); err != nil {
		log.Printf("Error validate courier: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		h.createErrorResponse(err.Error(), w)
		return
	}
	ctx := r.Context()

	courier, err := h.courierService.GetCourierWithLatestPosition(ctx, getCourierPayload.CourierId)
	if err != nil {
		log.Printf("Error getting courier: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		h.createErrorResponse("Error getting courier.", w)
		return
	}

	err = json.NewEncoder(w).Encode(courier)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		log.Printf("failed to encode json response: %v\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

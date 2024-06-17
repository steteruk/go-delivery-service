package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	nethttp "net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ErrDecodeFailed we return this error when we can not decode payload from http query.
var ErrDecodeFailed = errors.New("failed to decode payload")

// ErrValidatePayloadFailed throws this error when we have invalid payload.
var ErrValidatePayloadFailed = errors.New("failed to validated payload")

// ResponseMessage returns when we have bad request, or we have problem on server.
type ResponseMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type HandlerInterface interface {
	DecodePayloadFromJson(r *nethttp.Request, requestData any) error
	SuccessResponse(w nethttp.ResponseWriter, requestData any, status int)
	ValidatePayload(payload any) error
	FailResponse(w nethttp.ResponseWriter, errFailResponse error)
}

// Handler abstract handler we can reuse it in different handlers.
type Handler struct {
	Validator *validator.Validate
}

// DecodePayloadFromJson decodes payload from body http query and handle exceptions scenarios.
func (h *Handler) DecodePayloadFromJson(r *nethttp.Request, requestData any) error {
	err := json.NewDecoder(r.Body).Decode(requestData)

	if err != nil {
		log.Printf("incorrect json! please check your json formatting: %v\n", err)

		return ErrDecodeFailed
	}

	return nil
}

// SuccessResponse  Encodes response,that return user for http query and handle exceptions scenarios.
func (h *Handler) SuccessResponse(w nethttp.ResponseWriter, requestData any, status int) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(requestData)

	if err != nil {
		w.WriteHeader(nethttp.StatusInternalServerError)
		log.Panicf("failed to encode json response: %v\n", err)
	}

	w.WriteHeader(status)
}

// ValidatePayload validates some payload from http query.
func (h *Handler) ValidatePayload(payload any) error {
	err := h.Validator.Struct(payload)
	if err == nil {
		return nil
	}

	var errorMessage string
	if _, ok := err.(*validator.InvalidValidationError); ok {
		return fmt.Errorf("an invalid value was received: %w", err)
	}

	for _, errStruct := range err.(validator.ValidationErrors) {
		message := fmt.Sprintf("Incorrect Value %s %f", errStruct.StructField(), errStruct.Value())
		errorMessage += message + ","
	}

	errorMessage = strings.Trim(errorMessage, ",")

	return fmt.Errorf("%v:%w", errorMessage, ErrValidatePayloadFailed)
}

// FailResponse returns response for bad request.
func (h *Handler) FailResponse(w nethttp.ResponseWriter, errFailResponse error) {
	w.Header().Set("Content-Type", "application/json")
	switch true {
	case errors.Is(errFailResponse, ErrDecodeFailed):
		err := json.NewEncoder(w).Encode(&ResponseMessage{
			Status:  "Error",
			Message: errFailResponse.Error(),
		})

		if err != nil {
			log.Printf("failed to encode json response: %v\n", err)
		}

		w.WriteHeader(nethttp.StatusBadRequest)

	case errors.Is(errFailResponse, ErrValidatePayloadFailed):
		log.Printf("validate payload: %v", errFailResponse)

		json.NewEncoder(w).Encode(&ResponseMessage{
			Status:  "Error",
			Message: errFailResponse.Error(),
		})

		w.WriteHeader(nethttp.StatusBadRequest)

	default:
		log.Printf("Server error: %v\n", errFailResponse)
		w.WriteHeader(nethttp.StatusInternalServerError)
	}
}

// NewHandler creates http handler for handling http requests.
func NewHandler() *Handler {
	return &Handler{
		Validator: validator.New(),
	}
}

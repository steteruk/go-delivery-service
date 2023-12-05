package handler

import (
	"github.com/gorilla/mux"
	"github.com/steteruk/go-delivery-service/courier/domain"
	pkghttp "github.com/steteruk/go-delivery-service/pkg/http"
	"log"
	"net/http"
)

type CourierHandler struct {
	courierService domain.CourierServiceInterface
	httpHandler    pkghttp.HandlerInterface
}

func NewCourierHandler(courierService domain.CourierServiceInterface, handler pkghttp.HandlerInterface) *CourierHandler {
	return &CourierHandler{
		courierService: courierService,
		httpHandler:    handler,
	}
}

type CreateCourierPayload struct {
	Firstname string `json:"firstname" validate:"required,lte=40"`
}

type GetCourierPayload struct {
	CourierId string `json:"courier_id" validate:"required,uuid"`
}

func (h *CourierHandler) CreateCourierHandler(w http.ResponseWriter, r *http.Request) {
	var courierPayload CreateCourierPayload

	if err := h.httpHandler.DecodePayloadFromJson(r, &courierPayload); err != nil {
		h.httpHandler.FailResponse(w, err)

		return
	}

	if err := h.httpHandler.ValidatePayload(&courierPayload); err != nil {
		h.httpHandler.FailResponse(w, err)

		return
	}

	ctx := r.Context()
	courier, err := h.courierService.SaveNewCourier(
		ctx,
		&domain.Courier{
			FirstName:   courierPayload.Firstname,
			IsAvailable: true,
		},
	)

	if err != nil {
		log.Printf("Failed to save courier: %v", err)
		h.httpHandler.FailResponse(w, err)

		return
	}

	h.httpHandler.SuccessResponse(w, courier, http.StatusCreated)
}

func (h *CourierHandler) GetCourierHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ctx := r.Context()
	courierId := vars["courier_id"]
	courierResponse, err := h.courierService.GetCourierWithLatestPosition(ctx, courierId)

	if err != nil {
		log.Printf("failed to get courier: %v", err)
		h.httpHandler.FailResponse(w, err)

		return
	}

	h.httpHandler.SuccessResponse(w, courierResponse, http.StatusOK)
}

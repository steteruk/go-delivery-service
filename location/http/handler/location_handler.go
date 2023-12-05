package handler

import (
	"github.com/gorilla/mux"
	"github.com/steteruk/go-delivery-service/location/domain"
	pkghttp "github.com/steteruk/go-delivery-service/pkg/http"
	"log"
	"net/http"
	"time"
)

type ResponseMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// LocationPayload imagine payload from http query
type LocationPayload struct {
	Latitude  float64 `json:"latitude" validate:"required,latitude"`
	Longitude float64 `json:"longitude" validate:"required,longitude"`
}

type LocationHandler struct {
	courierService domain.CourierLocationServiceInterface
	httpHandler    pkghttp.HandlerInterface
}

func NewLocationHandler(courierService domain.CourierLocationServiceInterface, handler pkghttp.HandlerInterface) *LocationHandler {
	return &LocationHandler{
		courierService: courierService,
		httpHandler:    handler,
	}
}

func (h *LocationHandler) LatestLocationHandler(w http.ResponseWriter, r *http.Request) {
	var locationPayload LocationPayload

	if err := h.httpHandler.DecodePayloadFromJson(r, &locationPayload); err != nil {
		h.httpHandler.FailResponse(w, err)

		return
	}

	if err := h.httpHandler.ValidatePayload(&locationPayload); err != nil {
		h.httpHandler.FailResponse(w, err)

		return
	}

	vars := mux.Vars(r)
	courierId := vars["courier_id"]
	ctx := r.Context()
	courierLocation := &domain.CourierLocation{
		courierId,
		locationPayload.Latitude,
		locationPayload.Longitude,
		time.Now(),
	}

	err := h.courierService.SaveLatestCourierLocation(ctx, courierLocation)
	if err != nil {
		log.Printf("failed to store latest courier position: %v", err)

		h.httpHandler.FailResponse(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

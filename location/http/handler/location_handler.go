package handler

import (
	"github.com/gorilla/mux"
	"github.com/steteruk/go-delivery-service/location/domain"
	pkghttp "github.com/steteruk/go-delivery-service/pkg/http"
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
	courierLocationWorkerPool domain.CourierLocationWorkerPool
	httpHandler               pkghttp.HandlerInterface
}

func NewLocationHandler(courierLocationWorkerPool domain.CourierLocationWorkerPool, handler pkghttp.HandlerInterface) *LocationHandler {
	return &LocationHandler{
		courierLocationWorkerPool: courierLocationWorkerPool,
		httpHandler:               handler,
	}
}

// LatestLocationHandler handles request depending on location courier and validate query have valid payload and save data from payload in storage.
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
	courierLocation := &domain.CourierLocation{
		courierId,
		locationPayload.Latitude,
		locationPayload.Longitude,
		time.Now(),
	}

	h.courierLocationWorkerPool.AddTask(courierLocation)
	w.WriteHeader(http.StatusNoContent)
}

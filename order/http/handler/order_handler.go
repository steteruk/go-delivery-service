package handler

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/steteruk/go-delivery-service/order/domain"
	pkghttp "github.com/steteruk/go-delivery-service/pkg/http"
)

type OrderHandler struct {
	orderService domain.OrderService
	httpHandler  pkghttp.HandlerInterface
}

func NewOrderHandler(orderService domain.OrderService, handler pkghttp.HandlerInterface) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		httpHandler:  handler,
	}
}

type CreateOrderPayload struct {
	CustomerPhoneNumber string `json:"customer_phone_number" validate:"required,e164"`
}

type CreateOrderResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type GetOrderPayload struct {
	OrderID string `json:"order_id" validate:"required,uuid"`
}

type GetOrderResponse struct {
	Status string `json:"status"`
}

func (h *OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var orderPayload CreateOrderPayload

	if err := h.httpHandler.DecodePayloadFromJson(r, &orderPayload); err != nil {
		h.httpHandler.FailResponse(w, err)

		return
	}

	if err := h.httpHandler.ValidatePayload(&orderPayload); err != nil {
		h.httpHandler.FailResponse(w, err)

		return
	}

	ctx := r.Context()
	order := h.orderService.NewOrder(orderPayload.CustomerPhoneNumber)
	order, err := h.orderService.CreateOrder(
		ctx,
		order,
	)
	if err != nil {
		log.Printf("Failed to save order: %v", err)
		h.httpHandler.FailResponse(w, err)

		return
	}

	orderRes := &CreateOrderResponse{ID: order.ID, Status: order.Status}
	h.httpHandler.SuccessResponse(w, orderRes, http.StatusAccepted)
}

func (h *OrderHandler) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ctx := r.Context()
	orderPayload := &GetOrderPayload{OrderID: vars["order_id"]}
	if err := h.httpHandler.ValidatePayload(orderPayload); err != nil {
		h.httpHandler.FailResponse(w, err)

		return
	}

	order, err := h.orderService.GetOrderByID(ctx, orderPayload.OrderID)
	if err != nil {
		log.Printf("failed to get order: %v", err)
		h.httpHandler.FailResponse(w, err)

		return
	}

	orderRes := &GetOrderResponse{Status: order.Status}
	h.httpHandler.SuccessResponse(w, orderRes, http.StatusOK)
}

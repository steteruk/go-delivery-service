package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const OrderNewStatus = "pending"
const EventOrderCreated = "created"
const EventOrderUpdated = "updated"
const OrderStatusAccepted = "accepted"

// ErrOrderNotFound shows type this error, when we don't have order in db
var ErrOrderNotFound = errors.New("order was not found")
var ErrOrderValidationNotFound = errors.New("order validation was not found")

// CourierPayload gets from service courier data and need for unmarshal from payload object that have payloads field any
type CourierPayload struct {
	CourierID string `json:"courier_id"`
}

type Order struct {
	ID                  string    `json:"id"`
	CourierID           string    `json:"courier_id"`
	CustomerPhoneNumber string    `json:"customer_phone_number"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
}

// OrderValidation imagine entity for order validation for saving in db
type OrderValidation struct {
	OrderID            string
	CourierValidatedAt time.Time
	UpdatedAt          time.Time
	CourierError       string
}

// OrderValidationPayload imagine payload for order validation for different services
type OrderValidationPayload struct {
	CourierID string
}

type OrderRepository interface {
	SaveNewOrder(ctx context.Context, order *Order) (*Order, error)
	GetOrderByID(ctx context.Context, orderID string) (*Order, error)
	SaveOrderValidation(ctx context.Context, orderValidation *OrderValidation) error
	UpdateOrder(ctx context.Context, order *Order) error
	GetOrderValidationByID(ctx context.Context, orderID string) (*OrderValidation, error)
	UpdateOrderValidation(ctx context.Context, orderValidation *OrderValidation) error
}

type OrderService interface {
	GetOrderByID(ctx context.Context, orderID string) (*Order, error)
	CreateOrder(ctx context.Context, order *Order) (*Order, error)
	NewOrder(phoneNumber string) *Order
	ValidateOrderForService(ctx context.Context, serviceName string, orderID string, orderValidationPayload *OrderValidationPayload) error
}

type OrderServiceManager struct {
	orderRepo      OrderRepository
	orderPublisher OrderPublisher
}

// OrderPublisher publish message some systems.
type OrderPublisher interface {
	PublishOrder(ctx context.Context, order *Order, event string) error
}

func NewOrderService(orderRepo OrderRepository, publisher OrderPublisher) OrderService {
	return &OrderServiceManager{
		orderRepo:      orderRepo,
		orderPublisher: publisher,
	}
}

// CheckValidation checks validation for all services after that we change status order if order pass validation
func (orderValidation *OrderValidation) CheckValidation() bool {

	if orderValidation.CourierError != "" {
		return false
	}

	return true
}

func (s *OrderServiceManager) CreateOrder(ctx context.Context, order *Order) (*Order, error) {
	order, err := s.orderRepo.SaveNewOrder(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to store new order in the repository: %w", err)
	}
	err = s.orderPublisher.PublishOrder(ctx, order, EventOrderCreated)

	if err != nil {
		return nil, fmt.Errorf("failed to publish order: %w", err)
	}

	return order, nil
}

func (s *OrderServiceManager) GetOrderByID(ctx context.Context, orderID string) (*Order, error) {
	return s.orderRepo.GetOrderByID(ctx, orderID)
}

// NewOrder creates new order for saving in db
func (s *OrderServiceManager) NewOrder(phoneNumber string) *Order {
	return &Order{
		CustomerPhoneNumber: phoneNumber,
		CreatedAt:           time.Now(),
		Status:              OrderNewStatus,
	}
}

// ValidateOrderForService updates order status and creates or saves order validation
func (s *OrderServiceManager) ValidateOrderForService(ctx context.Context, serviceName string, orderID string, orderValidationPayload *OrderValidationPayload) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	orderValidation, err := s.orderRepo.GetOrderValidationByID(ctx, orderID)

	createNewOrderValidation := errors.Is(err, ErrOrderValidationNotFound)

	if err != nil && !createNewOrderValidation {
		return fmt.Errorf("failed to get order validation: %w", err)
	}

	if orderValidation == nil {
		orderValidation = &OrderValidation{}
		orderValidation.OrderID = orderID
	}

	var isCourierUpdateInOrder bool

	switch serviceName {
	case "courier":
		order.CourierID = orderValidationPayload.CourierID
		orderValidation.CourierValidatedAt = time.Now()
		isCourierUpdateInOrder = true
	}

	if createNewOrderValidation {
		err = s.orderRepo.SaveOrderValidation(
			ctx,
			orderValidation,
		)
	} else {
		err = s.orderRepo.UpdateOrderValidation(
			ctx,
			orderValidation,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to save order in database during validation: %w", err)
	}

	isOrderValidated := orderValidation.CheckValidation()
	if isOrderValidated {
		order.Status = OrderStatusAccepted
	}

	if isCourierUpdateInOrder || isOrderValidated {
		err = s.orderRepo.UpdateOrder(ctx, order)

		if err != nil {
			return fmt.Errorf("failed to order order in database during validation: %w", err)
		}

	}

	if isOrderValidated {
		err = s.orderPublisher.PublishOrder(ctx, order, EventOrderUpdated)

		if err != nil {
			return fmt.Errorf("failed to publish a order in the kafka: %w", err)
		}
	}

	return nil
}

package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/steteruk/go-delivery-service/avro/v1"
	"github.com/steteruk/go-delivery-service/order/domain"
	"log"
)

const OrderValidationsTopic = "order_validations.v1"

// OrderConsumerValidation consumes message order validation from kafka
type OrderConsumerValidation struct {
	orderService domain.OrderService
}

// OrderMessageValidation sends in third system for service information about order assign.
type OrderMessageValidation struct {
	IsSuccessful bool            `json:"is_successful"`
	Payload      json.RawMessage `json:"payload"`
	ServiceName  string          `json:"service_name"`
	OrderID      string          `json:"order_id"`
}

// NewOrderConsumerValidation creates order validation consumer
func NewOrderConsumerValidation(orderService domain.OrderService) *OrderConsumerValidation {
	orderConsumer := &OrderConsumerValidation{
		orderService: orderService,
	}

	return orderConsumer
}

// HandleJSONMessage Handle kafka message in json format
func (orderConsumerValidation *OrderConsumerValidation) HandleJSONMessage(ctx context.Context, message []byte) error {
	orderValidationMessage := avro.NewOrderValidationMessage()
	if err := orderValidationMessage.UnmarshalJSON(message); err != nil {
		log.Printf("failed to unmarshal Kafka message into order validation struct: %v\n", err)

		return nil
	}

	orderValidationPayload := domain.OrderValidationPayload{
		CourierID: orderValidationMessage.Payload.Courier_id.String,
	}
	err := orderConsumerValidation.orderService.ValidateOrderForService(
		ctx,
		orderValidationMessage.Service_name,
		orderValidationMessage.Order_id,
		&orderValidationPayload,
	)

	if err != nil {
		return fmt.Errorf("failed to validate order: %w", err)
	}

	return nil
}

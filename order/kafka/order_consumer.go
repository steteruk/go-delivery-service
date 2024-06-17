package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/steteruk/go-delivery-service/order/domain"
	"log"
)

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
func NewOrderConsumerValidation(
	orderService domain.OrderService,
) *OrderConsumerValidation {
	orderConsumer := &OrderConsumerValidation{
		orderService: orderService,
	}

	return orderConsumer
}

func (ocv *OrderConsumerValidation) HandleJSONMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	var orderMessageValidation OrderMessageValidation
	if err := json.Unmarshal(message.Value, &orderMessageValidation); err != nil {
		log.Printf("failed to unmarshal Kafka message into order validation struct: %v\n", err)

		return nil
	}

	err := ocv.orderService.ValidateOrderForService(
		ctx,
		orderMessageValidation.ServiceName,
		orderMessageValidation.OrderID,
		orderMessageValidation.Payload,
	)

	if err != nil {
		return fmt.Errorf("failed to validate order: %w", err)
	}

	return nil
}

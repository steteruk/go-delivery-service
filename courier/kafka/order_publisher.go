package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/steteruk/go-delivery-service/courier/domain"
	pkgkafka "github.com/steteruk/go-delivery-service/pkg/kafka"
)

const OrderTopicValidation = "order_validations"

// OrderValidationPublisher publisher for kafka
type OrderValidationPublisher struct {
	publisher *pkgkafka.Publisher
}

// CourierPayload need for send order message validation in kafka
type CourierPayload struct {
	CourierID string `json:"courier_id"`
}

// OrderMessageValidation sends in third system for service information about order assign.
type OrderMessageValidation struct {
	IsSuccessful bool           `json:"is_successful"`
	Payload      CourierPayload `json:"payload"`
	OrderID      string         `json:"order_id"`
	ServiceName  string         `json:"service_name"`
}

// NewOrderValidationPublisher creates new publisher and init
func NewOrderValidationPublisher(publisher *pkgkafka.Publisher) *OrderValidationPublisher {
	orderValidationPublisher := OrderValidationPublisher{
		publisher: publisher,
	}

	return &orderValidationPublisher
}

// PublishValidationResult sends order message in json format in Kafka.
func (orderPublisher *OrderValidationPublisher) PublishValidationResult(ctx context.Context, courierAssigment *domain.CourierAssignment) error {
	messageOrderValidation := OrderMessageValidation{
		IsSuccessful: true,
		ServiceName:  "courier",
		OrderID:      courierAssigment.OrderID,
		Payload: CourierPayload{
			CourierID: courierAssigment.CourierID,
		},
	}

	message, err := json.Marshal(messageOrderValidation)

	if err != nil {
		return fmt.Errorf("failed to marshal order message validation before sending Kafka event: %w", err)
	}

	err = orderPublisher.publisher.PublishMessage(ctx, message, []byte(courierAssigment.OrderID))

	if err != nil {
		return fmt.Errorf("failed to publish order event: %w", err)
	}

	return nil
}

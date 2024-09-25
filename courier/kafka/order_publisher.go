package kafka

import (
	"context"
	"fmt"
	"github.com/steteruk/go-delivery-service/avro/v1"

	"github.com/steteruk/go-delivery-service/courier/domain"
	pkgkafka "github.com/steteruk/go-delivery-service/pkg/kafka"
)

const OrderTopicValidation = "order_validations.v1"

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
	orderValidationMessage := avro.NewOrderValidationMessage()
	orderValidationMessage.Order_id = courierAssigment.OrderID
	orderValidationMessage.Service_name = "courier"
	orderValidationMessage.Is_successful = true
	orderValidationMessage.Payload.Courier_id.String = courierAssigment.CourierID
	orderValidationMessage.Payload.Courier_id.Null = nil

	message, err := orderValidationMessage.MarshalJSON()

	if err != nil {
		return fmt.Errorf("failed to marshal order message validation before sending Kafka event: %w", err)
	}

	schema := orderValidationMessage.Schema()
	err = orderPublisher.publisher.PublishMessage(ctx, message, []byte(courierAssigment.OrderID), schema)

	if err != nil {
		return fmt.Errorf("failed to publish order message validation event: %w", err)
	}

	return nil
}

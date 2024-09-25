package kafka

import (
	"context"
	"fmt"
	"github.com/steteruk/go-delivery-service/avro/v1"
	"github.com/steteruk/go-delivery-service/order/domain"
	pkgkafka "github.com/steteruk/go-delivery-service/pkg/kafka"
)

const OrderTopic = "orders.v1"

// OrderPublisher publisher for kafka
type OrderPublisher struct {
	publisher *pkgkafka.Publisher
}

// OrderPayload uses for embedding order id and phone customer
type OrderPayload struct {
	OrderID string `json:"id"`
}

// OrderMessage will publish, when order create.
type OrderMessage struct {
	OrderPayload OrderPayload `json:"payload"`
	Event        string       `json:"event"`
}

// NewOrderPublisher creates new publisher and init
func NewOrderPublisher(publisher *pkgkafka.Publisher) *OrderPublisher {
	orderPublisher := OrderPublisher{
		publisher: publisher,
	}

	return &orderPublisher
}

func (orderPublisher *OrderPublisher) PublishOrder(ctx context.Context, order *domain.Order, event string) error {
	orderMessage := avro.NewOrderMessage()
	orderMessage.Payload.Order_id = order.ID
	orderMessage.Event = event
	message, err := orderMessage.MarshalJSON()
	schema := orderMessage.Schema()

	if err != nil {
		return fmt.Errorf("failed to marshal order before sending Kafka event: %w", err)
	}

	err = orderPublisher.publisher.PublishMessage(ctx, message, []byte(order.ID), schema)

	if err != nil {
		return fmt.Errorf("failed to publish order event: %w", err)
	}

	return nil
}

package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/steteruk/go-delivery-service/order/domain"
	pkgkafka "github.com/steteruk/go-delivery-service/pkg/kafka"
)

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

func NewOrderPublisher(publisher *pkgkafka.Publisher) domain.OrderPublisher {
	return &OrderPublisher{publisher: publisher}
}

func (op *OrderPublisher) PublishOrder(ctx context.Context, order *domain.Order, event string) error {
	messageOrder := OrderMessage{
		Event: event,
		OrderPayload: OrderPayload{
			OrderID: order.ID,
		},
	}

	message, err := json.Marshal(messageOrder)

	if err != nil {
		return fmt.Errorf("failed to marshal order before sending Kafka event: %w", err)
	}

	err = op.publisher.PublishMessage(ctx, message, []byte(order.ID))

	if err != nil {
		return fmt.Errorf("failed to publish order event: %w", err)
	}

	return nil
}

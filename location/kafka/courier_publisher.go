package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/steteruk/go-delivery-service/location/domain"
)

const topic = "latest_position_courier"

type CourierPublisher struct {
	publisher sarama.AsyncProducer
}

func NewCourierPublisher(addr string) (*CourierPublisher, error) {
	brokers := []string{addr}
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewManualPartitioner
	config.Producer.RequiredAcks = sarama.WaitForLocal
	producer, err := sarama.NewAsyncProducer(brokers, config)

	if err != nil {
		return nil, fmt.Errorf("failed to create a new sarama async producer: %w", err)
	}

	return &CourierPublisher{publisher: producer}, nil
}

func (cp *CourierPublisher) PublishLatestCourierGeoPosition(ctx context.Context, courierLocation *domain.CourierLocation) error {
	message, err := json.Marshal(courierLocation)
	if err != nil {
		return fmt.Errorf("failed to marshal courier location before sending Kafka event: %w", err)
	}
	prepareMessage := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
	cp.publisher.Input() <- prepareMessage
	return nil
}

package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/steteruk/go-delivery-service/location/domain"
	"log"
)

type CourierLocationConsumer struct {
	courierLocationRepository domain.CourierLocationRepositoryInterface
}

func NewCourierLocationConsumer(
	courierLocationRepository domain.CourierLocationRepositoryInterface,
) *CourierLocationConsumer {
	return &CourierLocationConsumer{
		courierLocationRepository: courierLocationRepository,
	}
}

func (c *CourierLocationConsumer) HandleJSONMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	var courierLocation domain.CourierLocation

	if err := json.Unmarshal(message.Value, &courierLocation); err != nil {
		log.Printf("failed to unmarshal Kafka message into courier location struct: %v\n", err)

		return nil
	}

	err := c.courierLocationRepository.SaveLatestCourierGeoPosition(ctx, &courierLocation)

	if err != nil {
		return fmt.Errorf("failed to save a courier location in the repository: %w", err)
	}

	return nil
}

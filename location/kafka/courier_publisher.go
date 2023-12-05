package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/steteruk/go-delivery-service/location/domain"
	pkgkafka "github.com/steteruk/go-delivery-service/pkg/kafka"
)

type CourierPublisher struct {
	publisher *pkgkafka.Publisher
}

func NewCourierLocationPublisher(publisher *pkgkafka.Publisher) *CourierPublisher {
	return &CourierPublisher{publisher: publisher}
}

func (cp *CourierPublisher) PublishLatestCourierGeoPosition(courierLocation *domain.CourierLocation) error {
	message, err := json.Marshal(courierLocation)
	if err != nil {
		return fmt.Errorf("failed to marshal courier location before sending Kafka event: %w", err)
	}
	cp.publisher.PublishMessage(message)

	return nil
}

package redis

import (
	"context"
	"fmt"
	coreRedis "github.com/redis/go-redis/v9"
	"github.com/steteruk/go-delivery-service/location/domain"
)

const courierLatestCordsKey = "courier_latest_cord"

type CourierRepository struct {
	client *coreRedis.Client
}

func NewCourierRepository(client *coreRedis.Client) *CourierRepository {

	return &CourierRepository{
		client: client,
	}
}

func (r *CourierRepository) SaveLatestCourierGeoPosition(ctx context.Context, courierLocation *domain.CourierLocation) error {
	l := &coreRedis.GeoLocation{
		Longitude: courierLocation.Longitude,
		Latitude:  courierLocation.Latitude,
		Name:      courierLocation.CourierID,
	}

	if err := r.client.GeoAdd(ctx, courierLatestCordsKey, l).Err(); err != nil {
		return fmt.Errorf("failed to add courier geo location into redis: %w", err)
	}

	return nil
}

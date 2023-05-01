package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"strconv"
)

const courierLatestCordsKey = "courier_latest_cord"

type CourierRepository struct {
	client *redis.Client
}

func NewCourierRepository() *CourierRepository {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}

	db := os.Getenv("REDIS_DB")
	dbNumber, err := strconv.Atoi(db)
	if db == "" || err != nil {
		dbNumber = 0
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   dbNumber,
	})
	return &CourierRepository{
		client: rdb,
	}
}

func (a *CourierRepository) SaveLatestCourierGeoPosition(ctx context.Context, courierID string, latitude, longitude float64) error {
	l := &redis.GeoLocation{Longitude: longitude, Latitude: latitude, Name: courierID}

	if err := a.client.GeoAdd(ctx, courierLatestCordsKey, l).Err(); err != nil {
		return fmt.Errorf("failed to add courier geo location into redis: %w", err)
	}

	return nil
}

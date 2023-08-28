package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/steteruk/go-delivery-service/location/domain"
)

type CourierRepository struct {
	client *sql.DB
}

func NewCourierRepository(client *sql.DB) *CourierRepository {
	return &CourierRepository{
		client: client,
	}
}

func (r *CourierRepository) SaveLatestCourierGeoPosition(ctx context.Context, courierLocation *domain.CourierLocation) error {
	sqlStatement := "INSERT INTO courier_latest_cord (courier_id, latitude, longitude, created_at) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING"
	_, err := r.client.ExecContext(
		ctx,
		sqlStatement,
		courierLocation.CourierID,
		courierLocation.Latitude,
		courierLocation.Longitude,
		courierLocation.CreatedAt,
	)

	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("row couirier location was not saved: %w", err)
	}

	return nil
}

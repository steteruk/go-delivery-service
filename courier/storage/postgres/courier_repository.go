package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/steteruk/go-delivery-service/courier/domain"
)

type CourierRepository struct {
	client *sql.DB
}

func NewCourierRepository(client *sql.DB) *CourierRepository {
	return &CourierRepository{
		client: client,
	}
}

func (r *CourierRepository) SaveNewCourier(ctx context.Context, courier *domain.Courier) (*domain.Courier, error) {
	sqlStatement := "INSERT INTO courier (firstname) VALUES ($1) RETURNING courier_id, firstname, is_available"
	row := r.client.QueryRowContext(
		ctx,
		sqlStatement,
		courier.FirstName,
	)

	newCourier := domain.Courier{}
	err := row.Scan(&newCourier.ID, &newCourier.FirstName, &newCourier.IsAvailable)

	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("an error occurred while saving: %w", err)
	}

	return &newCourier, nil
}

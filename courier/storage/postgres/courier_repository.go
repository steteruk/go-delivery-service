package postgres

import (
	"context"
	"database/sql"
	"errors"
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
	err := row.Scan(&newCourier.Id, &newCourier.FirstName, &newCourier.IsAvailable)

	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("an error occurred while saving: %w", err)
	}

	return &newCourier, nil
}

func (r *CourierRepository) GetCourierById(ctx context.Context, courierId string) (*domain.Courier, error) {
	sqlStatement := "SELECT * FROM courier WHERE courier_id = $1"
	row := r.client.QueryRowContext(
		ctx,
		sqlStatement,
		courierId,
	)

	courier := domain.Courier{}
	err := row.Scan(&courier.Id, &courier.FirstName, &courier.IsAvailable)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrCourierNotFound
	}
	if err != nil {
		return nil, err
	}

	return &courier, nil
}

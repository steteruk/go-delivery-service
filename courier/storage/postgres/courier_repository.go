package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/steteruk/go-delivery-service/courier/domain"
	"hash/fnv"
	"log"
	"time"
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
	sqlStatement := "INSERT INTO couriers (firstname) VALUES ($1) RETURNING courier_id, firstname, is_available"
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
	sqlStatement := "SELECT * FROM couriers WHERE courier_id = $1"
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

// AssignOrderToCourier assigns a free courier to order. It runs a transaction and after finding an available courier it inserts a record into order_assignments table. In case of concurrent request and having a conflict it just does nothing and returns already assigned courier
func (repo *CourierRepository) AssignOrderToCourier(ctx context.Context, orderID string) (courierAssignment *domain.CourierAssignment, err error) {
	ctx = context.Background()
	tx, err := repo.client.BeginTx(ctx, nil)
	if err != nil {
		return
	}

	defer func(tx *sql.Tx) {
		if err != nil {
			errRollBack := tx.Rollback()
			if errRollBack != nil {
				log.Printf("failed to rolback transaction: %v\n", errRollBack)
			}

			return
		}

		err = tx.Rollback()

		if errors.Is(err, sql.ErrTxDone) {
			err = nil

			return
		}

		log.Printf("failed to rolback transaction: %v\n", err)

		return
	}(tx)

	_, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", repo.hashOrderID(orderID))
	if err != nil {
		return
	}
	query := "SELECT courier_id, order_id, created_at FROM order_assignments WHERE order_id=$1"
	row := tx.QueryRowContext(
		ctx,
		query,
		orderID,
	)

	courierAssignment = &domain.CourierAssignment{}
	err = row.Scan(&courierAssignment.CourierID, &courierAssignment.OrderID, &courierAssignment.CreatedAt)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return
	}

	if err == nil {
		return
	}

	query = "UPDATE couriers SET is_available = FALSE " +
		"where courier_id = (SELECT courier_id FROM couriers WHERE is_available = TRUE LIMIT 1 FOR UPDATE) RETURNING courier_id"
	row = tx.QueryRowContext(
		ctx,
		query,
	)

	var courierID string

	err = row.Scan(&courierID)

	if errors.Is(err, sql.ErrNoRows) {
		err = domain.ErrCourierNotFound
		return
	}

	if err != nil {
		return
	}

	query = "INSERT INTO order_assignments (order_id, courier_id, created_at) VALUES ($1, $2, $3)"

	courierAssignment.CourierID = courierID
	courierAssignment.OrderID = orderID
	courierAssignment.CreatedAt = time.Now()
	_, err = tx.ExecContext(
		ctx,
		query,
		courierAssignment.OrderID,
		courierAssignment.CourierID,
		courierAssignment.CreatedAt,
	)

	if err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		return
	}

	return
}

func (repo *CourierRepository) hashOrderID(orderID string) int64 {
	h := fnv.New64a()
	h.Write([]byte(orderID))
	return int64(h.Sum64())
}

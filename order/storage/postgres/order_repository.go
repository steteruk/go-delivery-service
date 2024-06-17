package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/steteruk/go-delivery-service/order/domain"
	"time"
)

type OrderRepository struct {
	client *sql.DB
}

func NewOrderRepository(client *sql.DB) *OrderRepository {
	return &OrderRepository{
		client: client,
	}
}

func (r *OrderRepository) SaveNewOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	sqlStatement := "INSERT INTO orders (customer_phone_number, status, created_at) VALUES ($1, $2, $3) RETURNING id, courier_id, customer_phone_number, status, created_at"
	row := r.client.QueryRowContext(
		ctx,
		sqlStatement,
		order.CustomerPhoneNumber,
		order.Status,
		order.CreatedAt,
	)

	newOrder := domain.Order{}
	var courierID sql.NullString
	err := row.Scan(&newOrder.ID, &courierID, &newOrder.CustomerPhoneNumber, &newOrder.Status, &newOrder.CreatedAt)
	order.CourierID = courierID.String

	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("an error occurred while saving: %w", err)
	}

	return &newOrder, nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, orderID string) (*domain.Order, error) {
	sqlStatement := "SELECT * FROM orders WHERE id = $1"
	row := r.client.QueryRowContext(
		ctx,
		sqlStatement,
		orderID,
	)

	order := domain.Order{}
	var courierID sql.NullString
	err := row.Scan(&order.ID, &courierID, &order.CustomerPhoneNumber, &order.Status, &order.CreatedAt)
	order.CourierID = courierID.String

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) AssignCourierToOrder(ctx context.Context, orderID string, courierID string) (*domain.Order, error) {
	query := "UPDATE orders SET courier_id = $1, status = $2 WHERE id = $3  RETURNING id, courier_id, customer_phone_number, status, created_at"
	row := r.client.QueryRowContext(
		ctx,
		query,
		courierID,
		orderID,
		"accepted",
	)
	order := domain.Order{}
	err := row.Scan(&order.ID, &courierID, &order.CustomerPhoneNumber, &order.Status, &order.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	return &order, nil
}

// GetOrderValidationByID GetOrderValidationValidationById gets order validation by id from db
func (repo *OrderRepository) GetOrderValidationByID(ctx context.Context, orderID string) (*domain.OrderValidation, error) {
	query := "SELECT order_id, courier_validated_at, courier_error, updated_at FROM order_validations WHERE order_id=$1"

	row := repo.client.QueryRowContext(
		ctx,
		query,
		orderID,
	)

	var orderValidation domain.OrderValidation

	err := row.Scan(&orderValidation.OrderID, &orderValidation.CourierValidatedAt, &orderValidation.CourierError, &orderValidation.UpdatedAt)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrOrderValidationNotFound
	}

	return &orderValidation, nil
}

// SaveOrderValidation creates or updates order validation
func (repo *OrderRepository) SaveOrderValidation(
	ctx context.Context,
	orderValidation *domain.OrderValidation,
) error {
	query := "INSERT INTO order_validations(order_id, courier_validated_at, courier_error) VALUES ($1, $2, $3)"
	_, err := repo.client.ExecContext(
		ctx,
		query,
		orderValidation.OrderID,
		orderValidation.CourierValidatedAt,
		orderValidation.CourierError,
	)

	return err
}

// UpdateOrderValidation updates order validation when order validation was added
func (repo *OrderRepository) UpdateOrderValidation(
	ctx context.Context,
	orderValidation *domain.OrderValidation,
) error {

	query := "UPDATE  order_validations SET courier_validated_at = $2, courier_error = $3, updated_at = $4 WHERE updated_at=$5 AND order_id=$1"
	result, err := repo.client.ExecContext(
		ctx,
		query,
		orderValidation.OrderID,
		orderValidation.CourierValidatedAt,
		orderValidation.CourierError,
		time.Now(),
		orderValidation.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rowAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowAffected > 0 {
		return nil
	}

	return domain.ErrOrderValidationNotFound
}

// UpdateOrder update order in db after get data from services.
func (repo *OrderRepository) UpdateOrder(ctx context.Context, order *domain.Order) error {
	query := "UPDATE orders SET status=$1, courier_id=$2 WHERE id = $3 RETURNING id, customer_phone_number, status, created_at, courier_id"
	_, err := repo.client.ExecContext(
		ctx,
		query,
		order.Status,
		order.CourierID,
		order.ID,
	)

	return err
}

package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrCourierNotFound = errors.New("courier was not found")

type Courier struct {
	Id          string `json:"id"`
	FirstName   string `json:"firstname"`
	IsAvailable bool   `json:"is_available"`
}

type CourierWithLatestPosition struct {
	Id             string            `json:"id"`
	FirstName      string            `json:"first_name"`
	IsAvailable    bool              `json:"is_available"`
	LatestPosition *LocationPosition `json:"latest_position"`
}
type CourierClient interface {
	GetLatestPosition(ctx context.Context, courierID string) (*LocationPosition, error)
}

type LocationPosition struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type CourierRepository interface {
	SaveNewCourier(ctx context.Context, courier *Courier) (*Courier, error)
	GetCourierById(ctx context.Context, courierId string) (*Courier, error)
	AssignOrderToCourier(ctx context.Context, orderID string) (CourierAssignment *CourierAssignment, err error)
}

type CourierServiceManager struct {
	courierClient            CourierClient
	courierRepository        CourierRepository
	orderValidationPublisher OrderValidationPublisher
}

// OrderValidationPublisher publish order validation message in queue for order service.
type OrderValidationPublisher interface {
	PublishValidationResult(ctx context.Context, courierAssignment *CourierAssignment) error
}

// CourierAssignment has order assign courier
type CourierAssignment struct {
	OrderID   string    `json:"order_id"`
	CourierID string    `json:"courier_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CourierService interface {
	GetCourierWithLatestPosition(ctx context.Context, courierId string) (*CourierWithLatestPosition, error)
	SaveNewCourier(ctx context.Context, courier *Courier) (*Courier, error)
	AssignOrderToCourier(ctx context.Context, orderID string) error
}

func NewCourierService(client CourierClient, repo CourierRepository, orderValidationPublisher OrderValidationPublisher) *CourierServiceManager {
	return &CourierServiceManager{
		courierClient:            client,
		courierRepository:        repo,
		orderValidationPublisher: orderValidationPublisher,
	}
}

func (s *CourierServiceManager) GetCourierWithLatestPosition(ctx context.Context, courierId string) (*CourierWithLatestPosition, error) {
	var locationPosition *LocationPosition

	courier, err := s.courierRepository.GetCourierById(ctx, courierId)
	if err != nil {
		return nil, fmt.Errorf("failed to get courier from the repository: %w", err)
	}
	resp, err := s.courierClient.GetLatestPosition(ctx, courierId)
	isErrCourierNotFound := errors.Is(err, ErrCourierNotFound)
	if err != nil && !isErrCourierNotFound {
		return nil, fmt.Errorf("failed to get courier: %w", err)
	}
	if resp != nil {
		locationPosition = &LocationPosition{
			Latitude:  resp.Latitude,
			Longitude: resp.Longitude,
		}
	}

	return &CourierWithLatestPosition{
		FirstName:      courier.FirstName,
		Id:             courier.Id,
		IsAvailable:    courier.IsAvailable,
		LatestPosition: locationPosition,
	}, nil
}

func (s *CourierServiceManager) SaveNewCourier(ctx context.Context, courier *Courier) (*Courier, error) {
	return s.courierRepository.SaveNewCourier(ctx, courier)
}

// AssignOrderToCourier assign order to courier and send message in queue
func (s *CourierServiceManager) AssignOrderToCourier(ctx context.Context, orderID string) error {

	courierAssigment, err := s.courierRepository.AssignOrderToCourier(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to save a courier assigments in the repository: %w", err)
	}

	err = s.orderValidationPublisher.PublishValidationResult(ctx, courierAssigment)

	if err != nil {
		return fmt.Errorf("failed to publish a order message validation in kafka: %w", err)
	}

	return nil
}

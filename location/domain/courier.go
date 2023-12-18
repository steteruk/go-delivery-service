package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrCourierLocationNotFound = errors.New("courier location was not found")

// CourierLocationServiceInterface saves courier position in storage.
type CourierLocationServiceInterface interface {
	SaveLatestCourierLocation(
		ctx context.Context,
		courierLocation *CourierLocation,
	) error
}

type CourierLocationWorkerPool interface {
	AddTask(courierLocation *CourierLocation)
}

// CourierService saves and publishes courier location
type CourierService struct {
	courierRepository CourierLocationRepositoryInterface
	courierPublisher  CourierLocationPublisherInterface
}

// CourierLocationRepositoryInterface saves latest location position courier in storage.
type CourierLocationRepositoryInterface interface {
	SaveLatestCourierGeoPosition(ctx context.Context, courierLocation *CourierLocation) error
}

// CourierRepositoryInterface gets latest position by uuid from storage.
type CourierRepositoryInterface interface {
	CourierLocationRepositoryInterface
	GetLatestPositionCourierById(ctx context.Context, courierID string) (*CourierLocation, error)
}

// CourierLocationPublisherInterface publish message some systems.
type CourierLocationPublisherInterface interface {
	PublishLatestCourierGeoPosition(courierLocation *CourierLocation) error
}

// NewCourierService creates model currier location with current data.
func NewCourierService(repo CourierLocationRepositoryInterface, publisher CourierLocationPublisherInterface) *CourierService {
	return &CourierService{
		courierRepository: repo,
		courierPublisher:  publisher,
	}
}

// CourierLocation provides information about coords courier.
type CourierLocation struct {
	CourierID string    `json:"courier_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
}

func (cs *CourierService) SaveLatestCourierLocation(ctx context.Context, courierLocation *CourierLocation) error {
	err := cs.courierRepository.SaveLatestCourierGeoPosition(ctx, courierLocation)
	if err != nil {
		return fmt.Errorf("failed to store latest courier location in the repository: %w", err)
	}
	err = cs.courierPublisher.PublishLatestCourierGeoPosition(courierLocation)

	if err != nil {
		return fmt.Errorf("failed to publish latest courier location: %w", err)
	}

	return nil
}

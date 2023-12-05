package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrCourierLocationNotFound = errors.New("courier location was not found")

type CourierLocationServiceInterface interface {
	SaveLatestCourierLocation(
		ctx context.Context,
		courierLocation *CourierLocation,
	) error
}

type CourierService struct {
	courierRepository CourierLocationRepositoryInterface
	courierPublisher  CourierLocationPublisherInterface
}

type CourierLocationRepositoryInterface interface {
	SaveLatestCourierGeoPosition(ctx context.Context, courierLocation *CourierLocation) error
}

type CourierRepositoryInterface interface {
	CourierLocationRepositoryInterface
	GetLatestPositionCourierById(ctx context.Context, courierID string) (*CourierLocation, error)
}

type CourierLocationPublisherInterface interface {
	PublishLatestCourierGeoPosition(courierLocation *CourierLocation) error
}

func NewCourierService(repo CourierLocationRepositoryInterface, publisher CourierLocationPublisherInterface) *CourierService {
	return &CourierService{
		courierRepository: repo,
		courierPublisher:  publisher,
	}
}

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

package domain

import (
	"context"
	"errors"
	"fmt"
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
type CourierClientInterface interface {
	GetLatestPosition(ctx context.Context, courierID string) (*LocationPosition, error)
}

type LocationPosition struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type CourierRepositoryInterface interface {
	SaveNewCourier(ctx context.Context, courier *Courier) (*Courier, error)
	GetCourierById(ctx context.Context, courierId string) (*Courier, error)
}

type CourierService struct {
	courierClient     CourierClientInterface
	courierRepository CourierRepositoryInterface
}

type CourierServiceInterface interface {
	GetCourierWithLatestPosition(ctx context.Context, courierId string) (*CourierWithLatestPosition, error)
	SaveNewCourier(ctx context.Context, courier *Courier) (*Courier, error)
}

func NewCourierService(client CourierClientInterface, repo CourierRepositoryInterface) *CourierService {
	return &CourierService{
		courierClient:     client,
		courierRepository: repo,
	}
}

func (s *CourierService) GetCourierWithLatestPosition(ctx context.Context, courierId string) (*CourierWithLatestPosition, error) {
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

func (s *CourierService) SaveNewCourier(ctx context.Context, courier *Courier) (*Courier, error) {
	return s.courierRepository.SaveNewCourier(ctx, courier)
}

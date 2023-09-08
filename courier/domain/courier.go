package domain

import (
	"context"
)

type Courier struct {
	ID          string `json:"id"`
	FirstName   string `json:"firstname"`
	IsAvailable bool   `json:"is_available"`
}

type CourierRepositoryInterface interface {
	SaveNewCourier(ctx context.Context, courier *Courier) (*Courier, error)
}

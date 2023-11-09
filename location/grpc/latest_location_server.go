package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/steteruk/go-delivery-service/location/domain"
	pb "github.com/steteruk/go-delivery-service/proto/generate/location/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LatestLocationServer struct {
	pb.UnimplementedCourierServer
	CourierRepository domain.CourierRepositoryInterface
}

func (ll *LatestLocationServer) GetCourierLatestPosition(ctx context.Context, req *pb.GetCourierLatestPositionRequest) (*pb.GetCourierLatestPositionResponse, error) {
	latestPosition, err := ll.CourierRepository.GetLatestPositionCourierById(ctx, req.CourierId)
	if err != nil {
		return nil, fmt.Errorf("impossible to get courier geo position: %w", err)
	}

	isErrCourierNotFound := err != nil && errors.Is(err, domain.ErrCourierLocationNotFound)
	if isErrCourierNotFound {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Position Not found: %v", err),
		)
	}
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Position Not found: %v", err),
		)
	}

	return &pb.GetCourierLatestPositionResponse{
		Latitude:  latestPosition.Latitude,
		Longitude: latestPosition.Longitude,
	}, nil
}

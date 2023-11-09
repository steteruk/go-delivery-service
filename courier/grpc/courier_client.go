package grpc

import (
	"context"
	"github.com/steteruk/go-delivery-service/courier/domain"
	pb "github.com/steteruk/go-delivery-service/proto/generate/location/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"log"
)

type CourierLocationPositionClient struct {
	courierClientGrpc pb.CourierClient
}

func NewCourierClient(locationConnection *grpc.ClientConn) *CourierLocationPositionClient {
	clientCourier := pb.NewCourierClient(locationConnection)
	return &CourierLocationPositionClient{
		courierClientGrpc: clientCourier,
	}
}
func (cl CourierLocationPositionClient) GetLatestPosition(ctx context.Context, courierId string) (*domain.LocationPosition, error) {
	courierLatestPositionResponse, err := cl.courierClientGrpc.GetCourierLatestPosition(ctx, &pb.GetCourierLatestPositionRequest{CourierId: courierId})
	code, ok := status.FromError(err)
	if ok && code.Code() == codes.NotFound {
		log.Printf("Not Found: %v\n", err)
		return nil, domain.ErrCourierNotFound
	}

	if err != nil {
		return nil, err
	}
	locationPosition := domain.LocationPosition{
		Latitude:  courierLatestPositionResponse.Latitude,
		Longitude: courierLatestPositionResponse.Longitude,
	}

	return &locationPosition, nil
}
func NewCourierConnection(courierGrpcAddress string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	return grpc.Dial(courierGrpcAddress, opts...)
}

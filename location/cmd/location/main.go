package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	coreRedis "github.com/redis/go-redis/v9"
	"github.com/steteruk/go-delivery-service/location/domain"
	"github.com/steteruk/go-delivery-service/location/env"
	server "github.com/steteruk/go-delivery-service/location/grpc"
	"github.com/steteruk/go-delivery-service/location/http"
	"github.com/steteruk/go-delivery-service/location/http/handler"
	"github.com/steteruk/go-delivery-service/location/kafka"
	"github.com/steteruk/go-delivery-service/location/storage/postgres"
	redisStorage "github.com/steteruk/go-delivery-service/location/storage/redis"
	pkghttp "github.com/steteruk/go-delivery-service/pkg/http"
	pkgkafka "github.com/steteruk/go-delivery-service/pkg/kafka"
	pb "github.com/steteruk/go-delivery-service/proto/generate/location/v1"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	config, err := env.GetConfig()
	ctx := context.Background()
	if err != nil {
		log.Printf("Failed to parse variable env: %v\n", err)
		return
	}

	publisher, err := pkgkafka.NewPublisher(config.AddrKafka, "latest_position_courier")
	if err != nil {
		log.Printf("failed to create publisher: %v\n", err)
		return
	}
	courierLocationPublisher := kafka.NewCourierLocationPublisher(publisher)

	clientRedis := coreRedis.NewClient(&coreRedis.Options{
		Addr: config.AddrRedis,
		DB:   config.NumberDbRedis,
	})
	defer clientRedis.Close()
	repoRedis := redisStorage.NewCourierRepository(clientRedis)

	courierService := domain.NewCourierService(repoRedis, courierLocationPublisher)

	credsDb := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", config.DbUser, config.DbPassword, config.DbName)
	dbClient, err := sql.Open("postgres", credsDb)
	if err != nil {
		log.Fatalf("Error connection database: %v\n", err)
	}
	defer dbClient.Close()
	repoPostgres := postgres.NewCourierRepository(dbClient)

	var wg sync.WaitGroup
	wg.Add(2)
	go runHttpServer(ctx, config, &wg, courierService)
	go runGrpc(ctx, config, &wg, repoPostgres)
	wg.Wait()
}

func runHttpServer(ctx context.Context, config env.Config, wg *sync.WaitGroup, courierService domain.CourierLocationServiceInterface) {
	locationHandler := handler.NewLocationHandler(courierService, pkghttp.NewHandler())
	var courierLocationURL = fmt.Sprintf(
		"/courier/{courier_id:%s}/location",
		"[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}",
	)
	routes := map[string]pkghttp.Route{
		courierLocationURL: {
			Handler: locationHandler.LatestLocationHandler,
			Method:  "POST",
		},
	}

	router := pkghttp.NewRoute(routes, mux.NewRouter())
	http.ServerRun(ctx, router, config.PortServer)
	wg.Done()
}

func runGrpc(ctx context.Context, config env.Config, wg *sync.WaitGroup, courierRepo domain.CourierRepositoryInterface) {
	lis, err := net.Listen("tcp", config.CourierLatestPositionGrpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	courierLocationServer := grpc.NewServer()
	pb.RegisterCourierServer(courierLocationServer, &server.LatestLocationServer{
		CourierRepository: courierRepo,
	})
	go func() {
		if err := courierLocationServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %s", err)
		}
	}()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-ctx.Done()
	stop()
	courierLocationServer.GracefulStop()
	wg.Done()
}

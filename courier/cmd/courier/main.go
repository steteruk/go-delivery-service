package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/steteruk/go-delivery-service/courier/domain"
	"github.com/steteruk/go-delivery-service/courier/env"
	courierGrpc "github.com/steteruk/go-delivery-service/courier/grpc"
	"github.com/steteruk/go-delivery-service/courier/http/handler"
	"github.com/steteruk/go-delivery-service/courier/kafka"
	"github.com/steteruk/go-delivery-service/courier/storage/postgres"
	pkghttp "github.com/steteruk/go-delivery-service/pkg/http"
	pkgkafka "github.com/steteruk/go-delivery-service/pkg/kafka"
)

func main() {
	config, err := env.GetConfig()
	if err != nil {
		log.Printf("failed to parse variable env: %v\n", err)
		return
	}

	credsPostgres := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", config.DBUser, config.DBPassword, config.DBName)
	client, err := sql.Open("postgres", credsPostgres)
	if err != nil {
		log.Panicf("unable to connect to database: %v\n", err)
	}
	defer client.Close()

	courierRepo := postgres.NewCourierRepository(client)

	publisher, err := pkgkafka.NewPublisher(config.AddrKafka, kafka.OrderTopicValidation)
	if err != nil {
		log.Panicf("failed to create publisher: %v\n", err)
	}
	orderValidationPublisher := kafka.NewOrderValidationPublisher(publisher)

	courierGrpcConn, err := courierGrpc.NewCourierConnection(config.CourierGrpcPort)
	if err != nil {
		log.Panicf("error courier gRPC client connection: %v\n", err)
	}
	defer courierGrpcConn.Close()

	courierClient := courierGrpc.NewCourierClient(courierGrpcConn)
	courierService := domain.NewCourierService(courierClient, courierRepo, orderValidationPublisher)
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	defer stop()

	wg.Add(2)
	go runHttpServer(ctx, config, &wg, courierService)
	go runOrderConsumer(ctx, config, &wg, courierService)
	wg.Wait()
}

func runHttpServer(ctx context.Context, config env.Config, wg *sync.WaitGroup, courierService domain.CourierService) {
	courierHandler := handler.NewCourierHandler(courierService, pkghttp.NewHandler())
	courierLatestPositionURL := fmt.Sprintf(
		"/couriers/{courier_id:%s}",
		"[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}",
	)

	routes := map[string]pkghttp.Route{
		"/couriers": {
			Handler: courierHandler.CreateCourierHandler,
			Method:  "POST",
		},
		courierLatestPositionURL: {
			Handler: courierHandler.GetCourierHandler,
			Method:  "GET",
		},
	}

	router := pkghttp.NewRoute(routes, mux.NewRouter())
	pkghttp.ServerRun(ctx, router, config.PortServer)
	wg.Done()
}

func runOrderConsumer(ctx context.Context, config env.Config, wg *sync.WaitGroup, courierService domain.CourierService) {
	defer wg.Done()
	orderConsumer := kafka.NewOrderConsumer(courierService)
	consumer, err := pkgkafka.NewConsumer(
		orderConsumer,
		config.AddrKafka,
		config.Verbose,
		config.Oldest,
		config.Assignor,
		kafka.OrderTopic,
	)

	if err != nil {
		log.Panicf("Failed to create kafka consumer group: %v\n", err)
	}

	err = consumer.ConsumeMessage(ctx)

	if err != nil {
		log.Panicf("Failed to consume message: %v\n", err)
	}
}

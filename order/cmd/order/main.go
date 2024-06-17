package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/steteruk/go-delivery-service/order/domain"
	"github.com/steteruk/go-delivery-service/order/env"
	"github.com/steteruk/go-delivery-service/order/http/handler"
	"github.com/steteruk/go-delivery-service/order/kafka"
	"github.com/steteruk/go-delivery-service/order/storage/postgres"
	pkghttp "github.com/steteruk/go-delivery-service/pkg/http"
	pkgkafka "github.com/steteruk/go-delivery-service/pkg/kafka"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	config, err := env.GetConfig()
	if err != nil {
		log.Printf("failed to parse variable env: %v\n", err)
		return
	}

	credsPostgres := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", config.DBUser, config.DBPassword, config.DBName)
	clientPostgres, err := sql.Open("postgres", credsPostgres)
	if err != nil {
		log.Panicf("unable to connect to database: %v\n", err)
	}
	defer clientPostgres.Close()

	orderRepo := postgres.NewOrderRepository(clientPostgres)
	publisher, err := pkgkafka.NewPublisher(config.AddrKafka, "orders")
	if err != nil {
		log.Printf("failed to create publisher: %v\n", err)
		return
	}
	orderPublisher := kafka.NewOrderPublisher(publisher)

	orderService := domain.NewOrderService(orderRepo, orderPublisher)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	var wg sync.WaitGroup
	wg.Add(2)
	go runHttpServer(ctx, config, &wg, orderService)
	go runOrderConsumer(ctx, config, &wg, orderService)
	wg.Wait()

}

func runHttpServer(ctx context.Context, config env.Config, wg *sync.WaitGroup, orderService domain.OrderService) {
	orderHandler := handler.NewOrderHandler(orderService, pkghttp.NewHandler())

	defer wg.Done()
	routes := map[string]pkghttp.Route{
		"/orders": {
			Handler: orderHandler.CreateOrderHandler,
			Method:  "POST",
		},
		"/orders/{order_id}": {
			Handler: orderHandler.GetOrderHandler,
			Method:  "GET",
		},
	}

	router := pkghttp.NewRoute(routes, mux.NewRouter())
	pkghttp.ServerRun(ctx, router, config.PortServer)
}

func runOrderConsumer(ctx context.Context, config env.Config, wg *sync.WaitGroup, orderService domain.OrderService) {
	defer wg.Done()
	orderConsumer := kafka.NewOrderConsumerValidation(orderService)
	consumer, err := pkgkafka.NewConsumer(
		orderConsumer,
		config.AddrKafka,
		config.Verbose,
		config.Oldest,
		config.Assignor,
		"order_validations",
	)

	if err != nil {
		log.Panicf("Failed to create kafka consumer group: %v\n", err)
	}

	err = consumer.ConsumeMessage(ctx)

	if err != nil {
		log.Panicf("Failed to consume message: %v\n", err)
	}
}

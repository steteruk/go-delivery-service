package main

import (
	"github.com/steteruk/go-delivery-service/location/domain"
	"github.com/steteruk/go-delivery-service/location/env"
	"github.com/steteruk/go-delivery-service/location/http"
	"github.com/steteruk/go-delivery-service/location/http/handler"
	"github.com/steteruk/go-delivery-service/location/kafka"
	redisStorage "github.com/steteruk/go-delivery-service/location/storage/redis"
	"log"
)

func main() {
	config, err := env.GetConfig()
	if err != nil {
		log.Printf("Failed to parse variable env: %v\n", err)
		return
	}

	publisher, err := kafka.NewCourierPublisher(config.AddrKafka)
	if err != nil {
		log.Printf("failed to create publisher: %v\n", err)
		return
	}
	repo := redisStorage.NewCourierRepository(config.AddrRedis, config.NumberDbRedis)
	courierService := domain.NewCourierService(repo, publisher)

	locationHandler := handler.NewLocationHandler(courierService)
	locationRouter := http.NewRouter(locationHandler.CourierHandler).Init()
	if err := http.ServerRun(locationRouter); err != nil {
		log.Printf("Failed to run http server: %v", err)
	}
}

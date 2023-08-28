package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/steteruk/go-delivery-service/location/env"
	"github.com/steteruk/go-delivery-service/location/kafka"
	"github.com/steteruk/go-delivery-service/location/storage/postgres"
	"log"
)

func main() {
	config, err := env.GetConfig()
	if err != nil {
		log.Panicf("failed to parse variable env: %v\n", err)
	}
	ctx := context.Background()
	credsPostgres := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", config.DbUser, config.DbPassword, config.DbName)
	client, err := sql.Open("postgres", credsPostgres)
	if err != nil {
		log.Panicf("Unable to connect to database: %v\n", err)
	}
	defer client.Close()

	courierRepo := postgres.NewCourierRepository(client)
	consumer, err := kafka.NewCourierConsumer(courierRepo, config.AddrKafka, config.Verbose, config.Oldest, config.Assignor)
	if err != nil {
		log.Panicf("Failed to create kafka consumer group: %v\n", err)
	}
	err = consumer.ConsumeCourierLatestCourierGeoPositionMessage(ctx)
	if err != nil {
		log.Panicf("Failed to consume message: %v\n", err)
	}

}

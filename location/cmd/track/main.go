package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/steteruk/go-delivery-service/location/env"
	"github.com/steteruk/go-delivery-service/location/kafka"
	"github.com/steteruk/go-delivery-service/location/storage/postgres"
	pkgkafka "github.com/steteruk/go-delivery-service/pkg/kafka"
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

	courierLocationConsumer := kafka.NewCourierLocationConsumer(courierRepo)
	consumer, err := pkgkafka.NewConsumer(
		courierLocationConsumer,
		config.KafkaAddress,
		config.Verbose,
		config.Oldest,
		config.Assignor,
		kafka.LatestPositionCourierTopic,
		[]string{config.KafkaSchemaRegistryAddress},
	)

	if err != nil {
		log.Panicf("Failed to create kafka consumer group: %v\n", err)
	}

	err = consumer.ConsumeMessage(ctx)

	if err != nil {
		log.Panicf("Failed to consume message: %v\n", err)
	}
}

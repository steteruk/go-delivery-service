package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/steteruk/go-delivery-service/courier/env"
	"github.com/steteruk/go-delivery-service/courier/http"
	"github.com/steteruk/go-delivery-service/courier/http/handler"
	"github.com/steteruk/go-delivery-service/courier/storage/postgres"
	"log"
)

func main() {
	config, err := env.GetConfig()
	if err != nil {
		log.Printf("Failed to parse variable env: %v\n", err)
		return
	}

	credsPostgres := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", config.DbUser, config.DbPassword, config.DbName)
	client, err := sql.Open("postgres", credsPostgres)
	if err != nil {
		log.Panicf("Unable to connect to database: %v\n", err)
	}
	defer client.Close()

	courierRepo := postgres.NewCourierRepository(client)
	courierHandler := handler.NewCourierHandler(courierRepo)
	courierRouter := http.NewRouter()
	courierRouter.AddRoute("POST", "/couriers", courierHandler.CourierHandler)

	if err := http.ServerRun(courierRouter.Init(), config.PortServer); err != nil {
		log.Printf("Failed to run http server: %v", err)
	}
}

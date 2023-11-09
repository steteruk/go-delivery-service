package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/steteruk/go-delivery-service/courier/domain"
	"github.com/steteruk/go-delivery-service/courier/env"
	courierGrpc "github.com/steteruk/go-delivery-service/courier/grpc"
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

	courierGrpcConn, err := courierGrpc.NewCourierConnection(config.CourierGrpcPort)
	if err != nil {
		log.Panicf("Error Courier Server Connection: %v\n", err)
	}
	defer courierGrpcConn.Close()

	courierClient := courierGrpc.NewCourierClient(courierGrpcConn)
	courierService := domain.NewCourierService(courierClient, courierRepo)

	courierHandler := handler.NewCourierHandler(courierService)
	courierRouter := http.NewRouter()
	courierRouter.AddRoute("POST", "/couriers", courierHandler.SaveNewCourierHandler)
	courierRouter.AddRoute("GET", "/couriers/{courier_id}", courierHandler.GetCourierHandler)

	if err := http.ServerRun(courierRouter.Init(), config.PortServer); err != nil {
		log.Printf("Failed to run http server: %v", err)
	}
}

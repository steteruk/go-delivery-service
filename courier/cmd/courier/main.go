package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/steteruk/go-delivery-service/courier/domain"
	"github.com/steteruk/go-delivery-service/courier/env"
	courierGrpc "github.com/steteruk/go-delivery-service/courier/grpc"
	"github.com/steteruk/go-delivery-service/courier/http"
	"github.com/steteruk/go-delivery-service/courier/http/handler"
	"github.com/steteruk/go-delivery-service/courier/storage/postgres"
	pkghttp "github.com/steteruk/go-delivery-service/pkg/http"
	"log"
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

	courierGrpcConn, err := courierGrpc.NewCourierConnection(config.CourierGrpcPort)
	if err != nil {
		log.Panicf("error courier gRPC client connection: %v\n", err)
	}
	defer courierGrpcConn.Close()

	courierClient := courierGrpc.NewCourierClient(courierGrpcConn)
	courierService := domain.NewCourierService(courierClient, courierRepo)

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

	if err := http.ServerRun(router, config.PortServer); err != nil {
		log.Printf("failed to run http server: %v", err)
	}
}

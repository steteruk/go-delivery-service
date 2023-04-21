package main

import (
	server "github.com/steteruk/go-delivery-service/location/http"
	"log"
)

func main() {
	if err := server.Run(); err != nil {
		log.Printf("Failed to run http server: %v", err)
	}
}

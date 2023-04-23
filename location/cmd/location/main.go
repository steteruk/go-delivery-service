package main

import (
	"github.com/steteruk/go-delivery-service/location/http"
	"log"
)

func main() {
	if err := http.ServerRun(); err != nil {
		log.Printf("Failed to run http server: %v", err)
	}
}

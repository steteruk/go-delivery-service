package http

import (
	"fmt"
	"log"
	"net/http"
)

const serverPort = ":8888"

func Run() {
	http.Handle("/", GetRouter())
	fmt.Printf("Starting server at port %s\n", serverPort)
	if err := http.ListenAndServe(serverPort, nil); err != nil {
		log.Fatal(err)
	}
}

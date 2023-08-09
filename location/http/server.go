package http

import (
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

func ServerRun(locationRouter *mux.Router) error {
	port := ":8888"
	if os.Getenv("HTTP_PORT") != "" {
		port = os.Getenv("HTTP_PORT")
	}
	http.Handle("/", locationRouter)

	if err := http.ListenAndServe(port, nil); err != nil {
		return err
	}
	return nil
}

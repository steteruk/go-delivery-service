package http

import (
	"github.com/gorilla/mux"
	"net/http"
)

func ServerRun(locationRouter *mux.Router, port string) error {
	http.Handle("/", locationRouter)
	if err := http.ListenAndServe(port, nil); err != nil {
		return err
	}
	return nil
}

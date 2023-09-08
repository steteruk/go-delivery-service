package http

import (
	"github.com/gorilla/mux"
	"net/http"
)

func ServerRun(courierRouter *mux.Router, port string) error {
	http.Handle("/", courierRouter)
	if err := http.ListenAndServe(port, nil); err != nil {
		return err
	}
	return nil
}

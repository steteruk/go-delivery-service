package http

import (
	"net/http"
	"os"
)

func Run() error {

	port := ":8888"
	if os.Getenv("HTTP_PORT") != "" {
		port = os.Getenv("HTTP_PORT")
	}
	http.Handle("/", NewRouter().Init())
	if err := http.ListenAndServe(port, nil); err != nil {
		return err
	}
	return nil
}

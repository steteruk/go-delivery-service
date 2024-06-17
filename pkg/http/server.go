package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func ServerRun(ctx context.Context, locationRouter *mux.Router, port string) {
	http.Handle("/", locationRouter)
	srv := &http.Server{
		Addr:    port,
		Handler: nil,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(timeoutCtx); err != nil {
		log.Fatalf("listen: %s\n", err)
	}

	fmt.Println("Stop Server signal")
}

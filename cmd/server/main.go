package main

import (
	"net/http"

	"github.com/rshafikov/alertme/internal/server/routers"
	"github.com/rshafikov/alertme/internal/server/storage"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	store := storage.NewMemStorage()
	metricsRouter := routers.NewMetricsRouter(store)

	mux := http.NewServeMux()
	mux.Handle("/update/", http.StripPrefix("/update", metricsRouter))

	return http.ListenAndServe(`:8080`, mux)
}

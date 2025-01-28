package main

import (
	"net/http"

	"github.com/rshafikov/alertme/internal/server/routers"
	"github.com/rshafikov/alertme/internal/server/services"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	store := services.NewMemStorage()
	metricsHandler := routers.NewMetricsHandler(store)

	mux := http.NewServeMux()
	mux.Handle("/update/", http.StripPrefix("/update", metricsHandler))

	return http.ListenAndServe(`:8080`, mux)
}

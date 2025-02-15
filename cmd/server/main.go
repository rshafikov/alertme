package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/alertme/internal/server/config"
	"github.com/rshafikov/alertme/internal/server/routers/metrics"
	"github.com/rshafikov/alertme/internal/server/storage"
	"net/http"
)

func main() {
	config.InitServerFlags()
	if err := runServer(); err != nil {
		panic(err)
	}
}

func runServer() error {
	s := storage.NewMemStorage()
	mR := metrics.NewMetricsRouter(s)

	r := chi.NewRouter()
	r.Mount("/", mR.Routes())

	fmt.Println("Listening on:", config.Address.String())
	return http.ListenAndServe(config.Address.String(), r)
}

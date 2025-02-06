package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/alertme/internal/server/routers"
	"github.com/rshafikov/alertme/internal/server/storage"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	s := storage.NewMemStorage()
	mR := routers.NewMetricsRouter(s)

	r := chi.NewRouter()
	r.Mount("/", mR.Routes())

	return http.ListenAndServe(`:8080`, r)
}

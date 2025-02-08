package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/alertme/internal/server"
	"github.com/rshafikov/alertme/internal/server/routers"
	"github.com/rshafikov/alertme/internal/server/storage"
	"net/http"
)

func main() {
	server.InitServerFlags()
	if err := runServer(); err != nil {
		panic(err)
	}
}

func runServer() error {
	s := storage.NewMemStorage()
	mR := routers.NewMetricsRouter(s)

	r := chi.NewRouter()
	r.Mount("/", mR.Routes())

	fmt.Println("Listening on:", server.Address.String())
	return http.ListenAndServe(server.Address.String(), r)
}

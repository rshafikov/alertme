package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/alertme/internal/server/config"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/routers/metrics"
	"github.com/rshafikov/alertme/internal/server/storage"
	"log"
	"net/http"
)

func main() {
	config.InitServerConfiguration()

	if err := logger.Initialize(config.LogLevel); err != nil {
		log.Fatal(err)
	}

	if err := runServer(); err != nil {
		log.Fatal(err)
	}
}

func runServer() error {
	memStorage := storage.NewMemStorage()
	fileSaver := storage.NewFileSaver(memStorage, config.FileStoragePath)

	if config.Restore {
		err := fileSaver.LoadStorage()
		if err != nil {
			return err
		}
	}

	if config.StoreInterval > 0 {
		err := fileSaver.SaveStorageWithInterval(config.StoreInterval)
		if err != nil {
			return err
		}
	}

	dbs, err := storage.NewDBStorage(config.DatabaseURL)
	if err != nil {
		return err
	}

	mR := metrics.NewMetricsRouter(memStorage, dbs)

	r := chi.NewRouter()
	r.Mount("/", mR.Routes())

	return http.ListenAndServe(config.Address.String(), r)
}

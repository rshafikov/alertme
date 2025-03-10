package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/alertme/internal/server/config"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/routers/metrics"
	"github.com/rshafikov/alertme/internal/server/storage"
	"go.uber.org/zap"
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
	fileSaver := storage.NewFileSaver(config.FileStoragePath)
	memStorage := storage.NewMemStorage()

	if config.Restore {
		loadErr := fileSaver.LoadStorage(memStorage)
		if loadErr != nil {
			logger.Log.Error("Failed to load metrics to storage:", zap.Error(loadErr))
		} else {
			logger.Log.Info("Metrics successfully loaded to storage")
		}
	}

	err := fileSaver.SaveStorageWithInterval(config.StoreInterval, memStorage)
	if err != nil {
		return err
	}

	mR := metrics.NewMetricsRouter(memStorage)

	r := chi.NewRouter()
	r.Mount("/", mR.Routes())

	logger.Log.Info("Listening on", zap.String("address", config.Address.String()))
	return http.ListenAndServe(config.Address.String(), r)
}

package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/alertme/internal/server/config"
	"github.com/rshafikov/alertme/internal/server/routers/metrics"
	"github.com/rshafikov/alertme/internal/server/storage"
	"net/http"
)

func main() {
	config.InitServerConfiguration()
	if err := runServer(); err != nil {
		panic(err)
	}
}

func runServer() error {
	fileSaver := storage.NewFileSaver(config.FileStoragePath)
	memStorage := storage.NewMemStorage()

	if config.Restore {
		loadErr := fileSaver.LoadStorage(memStorage)
		if loadErr != nil {
			config.Log.Errorf("Failed to load metrics to storage: %v", loadErr)
		} else {
			config.Log.Info("Metrics successfully loaded to storage")
		}
	}

	err := fileSaver.SaveStorageWithInterval(config.StoreInterval, memStorage)
	if err != nil {
		return err
	}

	mR := metrics.NewMetricsRouter(memStorage)

	r := chi.NewRouter()
	r.Mount("/", mR.Routes())

	config.Log.Infoln("Listening on", config.Address.String())
	return http.ListenAndServe(config.Address.String(), r)
}

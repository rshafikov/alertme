package main

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/alertme/internal/server/config"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/routers/metrics"
	"github.com/rshafikov/alertme/internal/server/storage"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
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
	metricsStorage := storage.NewMemStorage()
	fileSaver := storage.NewFileSaver(metricsStorage, config.FileStoragePath)

	if config.Restore {
		_ = fileSaver.LoadStorage(context.Background())
	}

	if config.StoreInterval > 0 {
		err := fileSaver.SaveStorageWithInterval(context.Background(), config.StoreInterval)
		if err != nil {
			return err
		}
	}

	mR := metrics.NewMetricsRouter(metricsStorage)

	if config.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		databaseStorage := storage.NewDBStorage(config.DatabaseURL)
		err := databaseStorage.BootStrap(ctx)

		select {
		case <-ctx.Done():
			logger.Log.Warn("database bootstrap timeout", zap.Error(ctx.Err()))
			err = storage.ErrDB
		default:
		}

		if err != nil {
			if errors.Is(err, storage.ErrDB) {
				logger.Log.Warn("database bootstrap failed, in-memory storage will be used", zap.Error(err))
			} else {
				return err
			}
		} else {
			mR = metrics.NewMetricsRouter(databaseStorage)
		}

	}

	r := chi.NewRouter()
	r.Mount("/", mR.Routes())

	return http.ListenAndServe(config.Address.String(), r)
}

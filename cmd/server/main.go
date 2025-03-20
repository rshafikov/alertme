package main

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/routers/metrics"
	"github.com/rshafikov/alertme/internal/server/settings"
	"github.com/rshafikov/alertme/internal/server/storage"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

func main() {
	settings.InitServerConfiguration()

	if err := logger.Initialize(settings.CONF.LogLevel); err != nil {
		log.Fatal(err)
	}

	if err := runServer(); err != nil {
		log.Fatal(err)
	}
}

func runServer() error {
	metricsStorage := storage.NewMemStorage()
	fileSaver := storage.NewFileSaver(metricsStorage, settings.CONF.FileStoragePath)

	if settings.CONF.Restore {
		if err := restoreStorage(&fileSaver); err != nil {
			logger.Log.Error("failed to load storage", zap.Error(err))
		}
	}

	if settings.CONF.StoreInterval > 0 {
		if err := fileSaver.SaveStorageWithInterval(context.Background(), settings.CONF.StoreInterval); err != nil {
			return err
		}
	}

	mR, err := setupDatabase(metricsStorage)
	if err != nil {
		return err
	}

	return startServer(mR)
}

func restoreStorage(fileSaver *storage.FileSaver) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return fileSaver.LoadStorage(ctx)
}

func setupDatabase(memStorage *storage.MemStorage) (*metrics.Router, error) {
	if settings.CONF.DatabaseURL == "" {
		return metrics.NewMetricsRouter(memStorage), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	databaseStorage := storage.NewDBStorage(settings.CONF.DatabaseURL)
	if err := databaseStorage.BootStrap(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Log.Warn("database bootstrap timeout", zap.Error(err))
			return metrics.NewMetricsRouter(memStorage), nil
		}
		if errors.Is(err, storage.ErrDB) {
			logger.Log.Warn("database bootstrap failed, in-memory storage will be used", zap.Error(err))
			return metrics.NewMetricsRouter(memStorage), nil
		}
		return nil, err
	}

	return metrics.NewMetricsRouter(databaseStorage), nil
}

func startServer(mR *metrics.Router) error {
	r := chi.NewRouter()
	r.Mount("/", mR.Routes())

	return http.ListenAndServe(settings.CONF.ServerAddress.String(), r)
}

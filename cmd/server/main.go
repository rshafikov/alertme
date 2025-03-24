package main

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/alertme/internal/server/database"
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
	memStorage := storage.NewMemStorage()
	fileSaver := storage.NewFileSaver(memStorage, settings.CONF.FileStoragePath)

	if settings.CONF.Restore {
		if err := restoreStorage(&fileSaver); err != nil {
			logger.Log.Error("failed to load storage", zap.Error(err))
		}
	}

	metricsRouter := metrics.NewMetricsRouter(memStorage)

	db, _ := setupDB(settings.CONF.DatabaseURL)
	if db != nil {
		metricsRouter = metrics.NewMetricsRouter(db)
	}

	if settings.CONF.StoreInterval > 0 && db == nil {
		if err := fileSaver.SaveStorageWithInterval(context.Background(), settings.CONF.StoreInterval); err != nil {
			return err
		}
	}

	return startServer(metricsRouter)
}

func restoreStorage(fileSaver *storage.FileSaver) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return fileSaver.LoadStorage(ctx)
}

func setupDB(dbURL string) (*database.DB, error) {
	if dbURL == "" {
		logger.Log.Info("database url not set, using in-memory database")
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.BootStrap(ctx, settings.CONF.DatabaseURL)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Log.Warn("database bootstrap timeout", zap.Error(err))
		}
		if errors.Is(err, database.ErrDB) || errors.Is(err, database.ErrConnToDB) {
			logger.Log.Warn("database bootstrap failed, in-memory storage will be used", zap.Error(err))
		}
		return nil, err
	}

	return db, nil
}

func startServer(mR *metrics.Router) error {
	r := chi.NewRouter()
	r.Mount("/", mR.Routes())

	return http.ListenAndServe(settings.CONF.ServerAddress.String(), r)
}

package storage

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/models"
	"go.uber.org/zap"
)

type DBStorage struct {
	URL  string
	Pool *pgxpool.Pool
}

func NewDBStorage(dbURL string) (*DBStorage, error) {
	var db DBStorage
	db.URL = dbURL

	if dbURL != "" {
		logger.Log.Info("using database:", zap.String("dbURL", dbURL))
		err := db.Connect()
		if err != nil {
			logger.Log.Error("failed to connect to database:", zap.Error(err))
			return nil, err
		}
		return &db, nil
	}

	return nil, errors.New(errmsg.URLCannotBeEmpty)
}

func (db *DBStorage) Connect() error {
	_, err := pgx.Connect(context.Background(), db.URL)

	if err != nil {
		logger.Log.Error("Unable to connect to database:", zap.Error(err))
		return err
	}

	db.Pool, err = pgxpool.New(context.Background(), db.URL)
	if err != nil {
		logger.Log.Error("Unable to connect DB pool:", zap.Error(err))
		return err
	}

	logger.Log.Debug("Connected to database")
	return nil
}

func (db *DBStorage) Add(m *models.Metric) error {
	return nil
}

func (db *DBStorage) Get(metricType models.MetricType, metricName string) (*models.Metric, error) {
	return nil, nil
}

func (db *DBStorage) List() []*models.Metric { return nil }

func (db *DBStorage) Clear() {}

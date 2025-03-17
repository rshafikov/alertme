package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/migrations"
	"github.com/rshafikov/alertme/internal/server/models"
	"go.uber.org/zap"
)

const updateQuery = `
	INSERT INTO metrics (name, value, delta, type)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (name) DO UPDATE
	SET value = EXCLUDED.value, delta = EXCLUDED.delta, type = EXCLUDED.type;
`

const getQuery = `
	SELECT name, value, delta, type
	FROM metrics
	WHERE type = $1 AND name = $2;
`

type DBStorage struct {
	URL  string
	Pool *pgxpool.Pool
}

//func (db *DBStorage) BootStrap(ctx context.Context)

func NewDBStorage(dbURL string) (*DBStorage, error) {
	var db DBStorage
	db.URL = dbURL

	if dbURL != "" {
		ctx := context.Background()
		err := db.Connect(ctx)
		if err != nil {
			return nil, err
		}
		err = db.MakeMigrations(ctx)
		if err != nil {
			return nil, err
		}
		logger.Log.Info("using database:", zap.String("dbURL", dbURL))
		return &db, nil
	}

	return nil, errors.New("url cannot be empty")
}

func (db *DBStorage) Connect(ctx context.Context) error {
	_, err := pgx.Connect(ctx, db.URL)

	if err != nil {
		logger.Log.Error("unable to connect to database", zap.Error(err))
		return err
	}

	db.Pool, err = pgxpool.New(ctx, db.URL)
	if err != nil {
		logger.Log.Error("unable to connect DB pool", zap.Error(err))
		return err
	}

	logger.Log.Debug("connected to database", zap.String("dbURL", db.URL))
	return nil
}

func (db *DBStorage) Add(ctx context.Context, m *models.Metric) error {
	if m.Type == models.CounterType {
		oldMetric, _ := db.Get(ctx, m.Type, m.Name)
		if oldMetric != nil {
			newDelta := *m.Delta + *oldMetric.Delta
			m.Delta = &newDelta
		}
	}

	_, err := db.Pool.Exec(ctx, updateQuery, m.Name, m.Value, m.Delta, m.Type)
	if err != nil {
		logger.Log.Error(errmsg.UnableToAddMetric, zap.Error(err))
		return err
	}

	logger.Log.Debug("metric added successfully", zap.String("name", m.Name))
	return nil
}

func (db *DBStorage) Get(ctx context.Context, metricType models.MetricType, metricName string) (*models.Metric, error) {
	var metric models.Metric
	err := db.Pool.QueryRow(ctx, getQuery, metricType, metricName).Scan(
		&metric.Name, &metric.Value, &metric.Delta, &metric.Type,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Log.Debug("metric not found", zap.String("name", metricName))
			return nil, err
		}
		logger.Log.Error("failed to get metric", zap.Error(err))
		return nil, err
	}

	logger.Log.Debug("metric retrieved successfully", zap.String("name", metricName))
	return &metric, nil
}

func (db *DBStorage) List(ctx context.Context) []*models.Metric {
	query := `
		SELECT name, value, delta, type
		FROM metrics;
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		logger.Log.Error("failed to list metrics", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var metrics []*models.Metric
	for rows.Next() {
		var metric models.Metric
		err := rows.Scan(&metric.Name, &metric.Value, &metric.Delta, &metric.Type)
		if err != nil {
			logger.Log.Error("failed to scan metric", zap.Error(err))
			continue
		}
		metrics = append(metrics, &metric)
	}

	if rows.Err() != nil {
		logger.Log.Error("error during rows iteration", zap.Error(rows.Err()))
	}

	logger.Log.Debug("metrics listed successfully", zap.Int("count", len(metrics)))
	return metrics
}

func (db *DBStorage) Clear(ctx context.Context) {
	query := `DELETE FROM metrics;`

	_, err := db.Pool.Exec(ctx, query)
	if err != nil {
		logger.Log.Error("failed to clear metrics", zap.Error(err))
		return
	}

	logger.Log.Debug("metrics cleared successfully")
}

func (db *DBStorage) AddBatch(ctx context.Context, metrics []*models.Metric) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, metric := range metrics {
		if metric.Type == models.CounterType {
			var oldDelta sql.NullInt64
			err = tx.QueryRow(
				ctx,
				"SELECT delta FROM metrics WHERE type = $1 AND name = $2",
				metric.Type, metric.Name,
			).Scan(&oldDelta)
			if err == nil && oldDelta.Valid {
				newDelta := *metric.Delta + oldDelta.Int64
				metric.Delta = &newDelta
			}
		}
		_, err = tx.Exec(ctx, updateQuery, metric.Name, metric.Value, metric.Delta, metric.Type)
		if err != nil {
			logger.Log.Error(errmsg.UnableToAddMetric, zap.Error(err))
			return err
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	logger.Log.Debug("metrics were added successfully")

	return nil
}

func (db *DBStorage) MakeMigrations(ctx context.Context) error {
	_, err := db.Pool.Exec(ctx, migrations.CreateMetricsType)
	if err != nil {
		logger.Log.Error("unable to create metric type enum", zap.Error(err))
		return err
	}

	_, err = db.Pool.Exec(ctx, migrations.CreateMetricsTable)
	if err != nil {
		logger.Log.Error("unable to make migrations", zap.Error(err))
		return err
	}
	return nil
}

func (db *DBStorage) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

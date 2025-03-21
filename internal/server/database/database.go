package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/migrations"
	"github.com/rshafikov/alertme/internal/server/models"
	"go.uber.org/zap"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

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

var ErrDB = errors.New("db error")

type DBConnection struct {
	URL  string
	Pool *pgxpool.Pool
}

func NewDBConnection(dbURL string) *DBConnection {
	return &DBConnection{
		URL:  dbURL,
		Pool: nil,
	}
}

func (c *DBConnection) Connect(ctx context.Context) error {
	_, err := pgx.Connect(ctx, c.URL)
	if err != nil {
		logger.Log.Error("unable to connect to database", zap.Error(err))
		return err
	}

	c.Pool, err = pgxpool.New(ctx, c.URL)
	if err != nil {
		logger.Log.Error("unable to connect DB pool", zap.Error(err))
		return err
	}

	logger.Log.Debug("connected to database", zap.String("dbURL", c.URL))
	return nil
}

func (c *DBConnection) Ping(ctx context.Context) error {
	return c.Pool.Ping(ctx)
}

type Migrator struct {
	Pool *pgxpool.Pool
}

func NewMigrator(pool *pgxpool.Pool) *Migrator {
	return &Migrator{Pool: pool}
}

func (m *Migrator) MakeMigrations(ctx context.Context) error {
	_, err := m.Pool.Exec(ctx, migrations.CreateMetricsType)
	if err != nil {
		logger.Log.Error("unable to create metric type enum", zap.Error(err))
		return err
	}

	_, err = m.Pool.Exec(ctx, migrations.CreateMetricsTable)
	if err != nil {
		logger.Log.Error("unable to make migrations", zap.Error(err))
		return err
	}
	return nil
}

type DB struct {
	Pool *pgxpool.Pool
}

func NewDB(pool *pgxpool.Pool) *DB {
	return &DB{Pool: pool}
}

func handlePgError(err error, warnMsg string, errorCode string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == errorCode {
		logger.Log.Warn(warnMsg)
		return ErrDB
	}
	return err
}

func BootStrap(ctx context.Context, dbURL string) (*DB, error) {
	conn := NewDBConnection(dbURL)
	if err := conn.Connect(ctx); err != nil {
		return nil, handlePgError(err, "unable to connect to database", pgerrcode.ConnectionException)
	}

	migrator := NewMigrator(conn.Pool)
	if err := migrator.MakeMigrations(ctx); err != nil {
		return nil, handlePgError(err, "schema already exists, skipping migrations", pgerrcode.DuplicateTable)
	}

	db := NewDB(conn.Pool)
	logger.Log.Info("using database:", zap.String("dbURL", dbURL))
	return db, nil
}

func (db *DB) Add(ctx context.Context, m *models.Metric) error {
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

func (db *DB) Get(ctx context.Context, metricType models.MetricType, metricName string) (*models.Metric, error) {
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

func (db *DB) List(ctx context.Context) []*models.Metric {
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

func (db *DB) Clear(ctx context.Context) {
	query := `DELETE FROM metrics;`

	_, err := db.Pool.Exec(ctx, query)
	if err != nil {
		logger.Log.Error("failed to clear metrics", zap.Error(err))
		return
	}

	logger.Log.Debug("metrics cleared successfully")
}

func (db *DB) AddBatch(ctx context.Context, metrics []*models.Metric) error {
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

func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

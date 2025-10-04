// Package database provides functionality for connecting to and interacting with a PostgreSQL database.
// It includes functions for establishing connections, running migrations, and performing CRUD operations on metrics.
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
	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/rshafikov/alertme/internal/server/retry"
	"go.uber.org/zap"
	"time"
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

const deleteAllQuery = `
	DELETE FROM metrics;
`

const getAllQuery = `
	SELECT name, value, delta, type
	FROM metrics;
`

// ErrDB is returned when there's an internal database error.
var ErrDB = errors.New("internal db error")

// ErrConnToDB is returned when the application cannot connect to the database.
var ErrConnToDB = errors.New("unable to connect to db")

// DBConnErrRetryIntervals defines the time intervals between retry attempts for database connection errors.
var DBConnErrRetryIntervals = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

// DB represents a database connection and provides methods for interacting with the database.
// It implements the Pinger interface.
type DB struct {
	// Pool is the connection pool used to execute database queries.
	Pool *pgxpool.Pool
}

// NewDB creates a new DB instance with the specified connection pool.
func NewDB(pool *pgxpool.Pool) *DB {
	return &DB{Pool: pool}
}

// handlePGErr processes PostgreSQL errors and returns appropriate error types.
// It logs the error with the provided warning message and checks if the error matches the expected error code.
// Returns ErrDB for matching PostgreSQL errors, ErrConnToDB for connection errors, or the original error.
func handlePGErr(err error, warnMsg, errorCode string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == errorCode {
		logger.Log.Debug(warnMsg, zap.Error(err))
		return ErrDB
	}

	var connErr *pgconn.ConnectError
	if errors.As(err, &connErr) {
		logger.Log.Debug(ErrConnToDB.Error(), zap.Error(err))
		return ErrConnToDB
	}

	if err != nil {
		logger.Log.Debug("unexpected non-pgerror happened", zap.Error(err))
		return err
	}

	return nil
}

// BootStrap sets up a database connection, runs migrations, and returns a DB instance.
// It connects to the database using the provided URL, applies migrations, and creates a new DB instance.
// Returns an error if the connection cannot be established or if migrations fail.
func BootStrap(ctx context.Context, dbURL string) (*DB, error) {
	conn := NewDBConnection(dbURL)
	if err := conn.Connect(ctx); err != nil {
		return nil, handlePGErr(err, "failed to establish connection with db", pgerrcode.ConnectionException)
	}

	migrator := NewMigrator(conn.Pool)
	if err := migrator.MakeMigrations(ctx); err != nil {
		return nil, handlePGErr(err, "schema already exists, skipping migrations", pgerrcode.DuplicateTable)
	}

	db := NewDB(conn.Pool)
	logger.Log.Info("using database:", zap.String("db_url", dbURL))
	return db, nil
}

// Add adds a metric to the database.
// For counter metrics, it adds the new delta to the existing delta if the metric already exists.
// It uses retry logic to handle database connection errors.
// Returns an error if the metric cannot be added.
func (db *DB) Add(ctx context.Context, m *models.Metric) error {
	if m.Type == models.CounterType {
		oldMetric, _ := db.Get(ctx, m.Type, m.Name)
		if oldMetric != nil {
			newDelta := *m.Delta + *oldMetric.Delta
			m.Delta = &newDelta
		}
	}

	if err := retry.OnErr(
		ctx,
		[]error{ErrConnToDB, ErrDB},
		DBConnErrRetryIntervals,
		func(args ...any) error {
			_, rawErr := db.Pool.Exec(ctx, updateQuery, m.Name, m.Value, m.Delta, m.Type)

			return handlePGErr(rawErr, "connection failed", pgerrcode.ConnectionException)
		},
	); err != nil {
		logger.Log.Error(errmsg.UnableToAddMetric, zap.Error(err))
		return err
	}
	logger.Log.Debug("metric added successfully", zap.String("name", m.Name))
	return nil
}

// Get retrieves a metric from the database by its type and name.
// It uses retry logic to handle database connection errors.
// Returns the metric if found, or an error if the metric doesn't exist or if there's a database error.
func (db *DB) Get(ctx context.Context, metricType models.MetricType, metricName string) (*models.Metric, error) {
	var metric models.Metric
	if err := retry.OnErr(
		ctx,
		[]error{ErrConnToDB, ErrDB},
		DBConnErrRetryIntervals,
		func(args ...any) error {
			rawErr := db.Pool.QueryRow(ctx, getQuery, metricType, metricName).Scan(
				&metric.Name, &metric.Value, &metric.Delta, &metric.Type)

			return handlePGErr(rawErr, "connection failed", pgerrcode.ConnectionException)
		},
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(errmsg.MetricNotFound)
		}
		return nil, err
	}

	return &metric, nil
}

// List retrieves all metrics from the database.
// It implements manual retry logic to handle database connection errors.
// Returns a slice of metrics, or nil if there's a database error.
func (db *DB) List(ctx context.Context) []*models.Metric {
	rows, err := db.Pool.Query(ctx, getAllQuery)

	err = handlePGErr(err, "failed to connect", pgerrcode.ConnectionException)
	for try, sleep := range []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second} {
		if errors.Is(err, ErrConnToDB) || errors.Is(err, ErrDB) {
			logger.Log.Debug("retrying to list metrics", zap.Int("retry", try+1), zap.Error(err))
			time.Sleep(sleep)
			rows, err = db.Pool.Query(ctx, getAllQuery)
			continue
		}
		break
	}
	defer rows.Close()

	if err != nil {
		logger.Log.Error("failed to list metrics", zap.Error(err))
		return nil
	}

	var metrics []*models.Metric
	for rows.Next() {
		var metric models.Metric
		scanErr := rows.Scan(&metric.Name, &metric.Value, &metric.Delta, &metric.Type)
		if scanErr != nil {
			logger.Log.Error("failed to scan metric", zap.Error(err))
			break
		}
		metrics = append(metrics, &metric)
	}

	if rows.Err() != nil {
		logger.Log.Error("error during rows iteration", zap.Error(rows.Err()))
	}

	logger.Log.Debug("metrics retrieved successfully", zap.Int("count", len(metrics)))
	return metrics
}

func (db *DB) Clear(ctx context.Context) {
	if err := retry.OnErr(
		ctx,
		[]error{ErrConnToDB, ErrDB},
		DBConnErrRetryIntervals,
		func(args ...any) error {
			_, rawErr := db.Pool.Exec(ctx, deleteAllQuery)
			return handlePGErr(rawErr, "connection failed", pgerrcode.ConnectionException)
		},
	); err != nil {
		logger.Log.Error("failed to clear metrics", zap.Error(err))
		return
	}
	logger.Log.Debug("metrics cleared successfully")
}

func (db *DB) AddBatch(ctx context.Context, metrics []*models.Metric) error {
	if err := retry.OnErr(
		ctx,
		[]error{ErrConnToDB, ErrDB},
		DBConnErrRetryIntervals,
		func(args ...any) error {
			rawErr := db.batchTx(ctx, metrics)
			return handlePGErr(rawErr, "connection failed", pgerrcode.ConnectionException)
		},
	); err != nil {
		logger.Log.Error("failed to add metrics", zap.Error(err))
		return err
	}

	return nil
}

func (db *DB) batchTx(ctx context.Context, metrics []*models.Metric) error {
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
	if err := retry.OnErr(
		ctx,
		[]error{ErrDB, ErrConnToDB},
		DBConnErrRetryIntervals,
		func(args ...any) error {
			return handlePGErr(db.Pool.Ping(ctx), "unable to ping db", pgerrcode.ConnectionException)
		},
	); err != nil {
		logger.Log.Error("failed to ping db", zap.Error(err))
		return err
	}

	return nil
}

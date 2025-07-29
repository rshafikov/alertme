package database

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/migrations"
	"go.uber.org/zap"
)

// Migrator handles database schema migrations.
// It applies SQL migrations to create or update database tables and types.
type Migrator struct {
	// Pool is the connection pool used to execute migration queries.
	Pool *pgxpool.Pool
}

// NewMigrator creates a new Migrator with the specified connection pool.
func NewMigrator(pool *pgxpool.Pool) *Migrator {
	return &Migrator{Pool: pool}
}

// MakeMigrations applies all migrations to the database.
// It creates the metric type enum and the metrics table.
// Returns an error if any migration fails.
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

package database

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/migrations"
	"go.uber.org/zap"
)

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

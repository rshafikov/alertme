package storage

import (
	"context"
	"github.com/rshafikov/alertme/internal/server/models"
)

type BaseMetricStorage interface {
	Add(ctx context.Context, metric *models.Metric) error
	Get(ctx context.Context, metricType models.MetricType, name string) (*models.Metric, error)
	List(ctx context.Context) []*models.Metric
	Clear(ctx context.Context)
	AddBatch(ctx context.Context, metrics []*models.Metric) error
}

type BaseMetricSaver interface {
	LoadStorage(ctx context.Context) error
	SaveStorage(ctx context.Context) error
	SaveMetrics(metrics []*models.Metric) error
	LoadMetrics() ([]*models.Metric, error)
}

type BaseDatabase interface {
	MakeMigrations(ctx context.Context) error
	Connect(ctx context.Context) error
	Ping(ctx context.Context) error
}

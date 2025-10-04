package storage

import (
	"context"
	"github.com/rshafikov/alertme/internal/server/models"
)

// BaseMetricStorage defines the interface for storing and retrieving metrics.
type BaseMetricStorage interface {
	// Add adds a single metric to the storage.
	Add(ctx context.Context, metric *models.Metric) error

	// Get retrieves a metric by its type and name.
	Get(ctx context.Context, metricType models.MetricType, name string) (*models.Metric, error)

	// List returns all metrics in the storage.
	List(ctx context.Context) []*models.Metric

	// Clear removes all metrics from the storage.
	Clear(ctx context.Context)

	// AddBatch adds multiple metrics to the storage in a single operation.
	AddBatch(ctx context.Context, metrics []*models.Metric) error
}

// BaseMetricSaver defines the interface for saving and loading metrics to/from persistent storage.
type BaseMetricSaver interface {
	// LoadStorage loads metrics from persistent storage into memory.
	LoadStorage(ctx context.Context) error

	// SaveStorage saves metrics from memory to persistent storage.
	SaveStorage(ctx context.Context) error

	// SaveMetrics saves the provided metrics to persistent storage.
	SaveMetrics(metrics []*models.Metric) error

	// LoadMetrics loads metrics from persistent storage and returns them.
	LoadMetrics() ([]*models.Metric, error)
}

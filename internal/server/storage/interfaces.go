package storage

import "github.com/rshafikov/alertme/internal/server/models"

type BaseMetricStorage interface {
	Add(metric *models.Metric) error
	Get(metricType models.MetricType, name string) (*models.Metric, error)
	List() []*models.Metric
	Clear()
}

type BaseMetricSaver interface {
	LoadStorage(storage BaseMetricStorage) error
	SaveStorage(storage BaseMetricStorage) error
	SaveMetrics(metrics []*models.Metric) error
	LoadMetrics() ([]*models.Metric, error)
}

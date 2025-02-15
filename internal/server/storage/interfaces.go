package storage

import "github.com/rshafikov/alertme/internal/server/models"

type BaseMetricStorage interface {
	Add(metric *models.Metric) error
	Get(metricType models.MetricType, name string) (models.Metric, error)
	List() ([]models.Metric, error)
	Clear()
}

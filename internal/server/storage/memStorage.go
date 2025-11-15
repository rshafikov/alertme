package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/models"
)

// MemStorage implements BaseMetricStorage interface using in-memory storage.
type MemStorage struct {
	metrics map[string]*models.Metric
	mu      sync.RWMutex
}

// NewMemStorage creates a new in-memory storage for metrics.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]*models.Metric),
	}
}

func (s *MemStorage) Add(ctx context.Context, m *models.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.addMetric(m)
}

func (s *MemStorage) Get(ctx context.Context, metricType models.MetricType, metricName string) (*models.Metric, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	metricMapName := fmt.Sprintf("%s-%s", metricType, metricName)
	metric, exists := s.metrics[metricMapName]
	if exists {
		return metric, nil
	}

	return nil, errors.New(errmsg.MetricNotFound)
}

func (s *MemStorage) List(ctx context.Context) []*models.Metric {
	s.mu.Lock()
	defer s.mu.Unlock()

	metrics := make([]*models.Metric, 0, len(s.metrics))
	for _, metric := range s.metrics {
		metrics = append(metrics, metric)
	}

	return metrics
}

func (s *MemStorage) Clear(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics = make(map[string]*models.Metric)
}

func (s *MemStorage) AddBatch(ctx context.Context, metrics []*models.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, metric := range metrics {
		if err := s.addMetric(metric); err != nil {
			return err
		}
	}
	return nil
}

func (s *MemStorage) addMetric(m *models.Metric) error {
	existingMetric, exists := s.metrics[m.MapName()]

	switch m.Type {
	case models.GaugeType:
		if m.Value == nil {
			return errors.New("metric gauge value cannot be nil")
		}
		s.metrics[m.MapName()] = m
	case models.CounterType:
		if m.Delta == nil {
			return errors.New("metric counter delta cannot be nil")
		}
		newDelta := *m.Delta
		if exists {
			newDelta += *existingMetric.Delta
		}
		s.metrics[m.MapName()] = &models.Metric{
			Name:  m.Name,
			Type:  m.Type,
			Delta: &newDelta,
		}
	default:
		return errors.New(errmsg.InvalidMetricType)
	}

	return nil
}

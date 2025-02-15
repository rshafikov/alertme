package storage

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/rshafikov/alertme/internal/server/models"
)

const (
	NotFoundMetricErrMsg        = "metric not found"
	UnsupportedMetricTypeErrMsg = "unsupported metric type"
	CannotConvertToIntErrMsg    = "cannot convert to int"
	CannotConvertToFloatErrMsg  = "cannot convert to float"
)

type UnsupportedMetricTypeError struct {
	arg     string
	message string
}

func (e *UnsupportedMetricTypeError) Error() string {
	return fmt.Sprintf("'%s' <- %s", e.arg, e.message)
}

type IncorrectMetricValueError struct {
	arg     string
	message string
}

func (e *IncorrectMetricValueError) Error() string {
	return fmt.Sprintf("'%s' <- %s", e.arg, e.message)
}

type MemStorage struct {
	Gauges   map[string]models.GaugeMetric
	Counters map[string]models.CounterMetric
	mu       sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]models.GaugeMetric),
		Counters: make(map[string]models.CounterMetric),
	}
}

func (s *MemStorage) Add(m *models.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch m.Type {
	case models.GaugeType:
		value, err := strconv.ParseFloat(m.Value, 64)
		if err != nil {
			return &IncorrectMetricValueError{arg: m.Value, message: CannotConvertToFloatErrMsg}
		}
		newGaugeMetric := models.GaugeMetric{Type: m.Type, Name: m.Name, Value: value}
		s.Gauges[m.Name] = newGaugeMetric
	case models.CounterType:
		value, err := strconv.ParseInt(m.Value, 10, 64)
		if err != nil {
			return &IncorrectMetricValueError{arg: m.Value, message: CannotConvertToIntErrMsg}
		}
		if oldMetric, ok := s.Counters[m.Name]; ok {
			value += oldMetric.Value
		}
		newCounterMetric := models.CounterMetric{Type: m.Type, Name: m.Name, Value: value}
		s.Counters[m.Name] = newCounterMetric
	default:
		return &UnsupportedMetricTypeError{arg: string(m.Type), message: UnsupportedMetricTypeErrMsg}
	}

	return nil
}

func (s *MemStorage) Get(metricType models.MetricType, name string) (models.Metric, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch metricType {
	case models.GaugeType:
		if metric, ok := s.Gauges[name]; ok {
			value := strconv.FormatFloat(metric.Value, 'f', -1, 64)
			return models.Metric{Type: metric.Type, Name: metric.Name, Value: value}, nil
		}
		return models.Metric{}, errors.New(NotFoundMetricErrMsg)
	case models.CounterType:
		if metric, ok := s.Counters[name]; ok {
			value := strconv.FormatInt(metric.Value, 10)
			return models.Metric{Type: metric.Type, Name: metric.Name, Value: value}, nil
		}
		return models.Metric{}, errors.New(NotFoundMetricErrMsg)
	default:
		return models.Metric{}, errors.New(UnsupportedMetricTypeErrMsg)
	}

}

func (s *MemStorage) List() ([]models.Metric, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	metrics := make([]models.Metric, 0, len(s.Gauges)+len(s.Counters))
	for _, metric := range s.Gauges {
		value := strconv.FormatFloat(metric.Value, 'f', -1, 64)
		metrics = append(metrics, models.Metric{Type: metric.Type, Name: metric.Name, Value: value})
	}
	for _, metric := range s.Counters {
		value := strconv.FormatInt(metric.Value, 10)
		metrics = append(metrics, models.Metric{Type: metric.Type, Name: metric.Name, Value: value})
	}
	return metrics, nil
}

func (s *MemStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Counters = make(map[string]models.CounterMetric)
	s.Gauges = make(map[string]models.GaugeMetric)
}

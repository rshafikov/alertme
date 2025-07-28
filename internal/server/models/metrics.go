package models

import (
	"errors"
	"fmt"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"strconv"
	"strings"
)

type MetricType string

const (
	GaugeType   MetricType = "gauge"
	CounterType MetricType = "counter"
)

type PlainMetric struct {
	Name  string
	Type  MetricType
	Value string
}

func (pm *PlainMetric) ConverToMetric() (*Metric, error) {
	switch pm.Type {
	case GaugeType:
		value, err := strconv.ParseFloat(pm.Value, 64)
		if err != nil {
			return nil, errors.New(errmsg.UnableToParseFloat)
		}
		return &Metric{Type: pm.Type, Name: pm.Name, Value: &value}, nil

	case CounterType:
		delta, err := strconv.ParseInt(pm.Value, 10, 64)
		if err != nil {
			return nil, errors.New(errmsg.UnableToParseInt)
		}
		return &Metric{Type: pm.Type, Name: pm.Name, Delta: &delta}, nil

	default:
		return nil, errors.New(errmsg.InvalidMetricType)
	}
}

type Metric struct {
	Name    string     `json:"id"`
	Value   *float64   `json:"value,omitempty"`
	Delta   *int64     `json:"delta,omitempty"`
	Type    MetricType `json:"type"`
	mapName string
}

func (m *Metric) String() string {
	switch m.Type {
	case GaugeType:
		return strings.TrimRight(fmt.Sprintf("%.*f", 7, *m.Value), "0")
	case CounterType:
		return fmt.Sprintf("%v", *m.Delta)
	default:
		return "unknown metric type"
	}
}

// MapName m.Name-m.Type
func (m *Metric) MapName() string {
	if m.mapName == "" {
		m.mapName = string(m.Type) + "-" + m.Name
	}
	return m.mapName
}

func (m *Metric) ConvertToPlain() *PlainMetric {
	return &PlainMetric{
		Name:  m.Name,
		Type:  m.Type,
		Value: m.String(),
	}
}

func NewMetric(metricType MetricType, metricName, metricValue string) (*Metric, error) {
	plainMetric := &PlainMetric{
		Name:  metricName,
		Type:  metricType,
		Value: metricValue,
	}
	return plainMetric.ConverToMetric()
}

package models

import (
	"fmt"
)

type MetricType string

const (
	GaugeType   MetricType = "gauge"
	CounterType MetricType = "counter"
)

type Metric struct {
	Value string     `json:"value"`
	Name  string     `json:"name"`
	Type  MetricType `json:"type"`
}

func (m Metric) String() string {
	return fmt.Sprintf("%s@%s:%s", m.Name, m.Type, m.Value)
}

type GaugeMetric struct {
	Value float64
	Name  string
	Type  MetricType
}

type CounterMetric struct {
	Value int64
	Name  string
	Type  MetricType
}

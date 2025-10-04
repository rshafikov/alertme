// Package models provides data structures and operations for working with metrics.
package models

import (
	"errors"
	"fmt"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"strconv"
	"strings"
)

// MetricType represents the type of a metric (gauge or counter).
type MetricType string

const (
	// GaugeType represents a metric that can arbitrarily go up and down.
	GaugeType MetricType = "gauge"
	// CounterType represents a metric that can only increase.
	CounterType MetricType = "counter"
)

// PlainMetric represents a metric with string values for serialization/deserialization.
type PlainMetric struct {
	// Name is the identifier of the metric.
	Name string
	// Type specifies whether this is a gauge or counter metric.
	Type MetricType
	// Value is the string representation of the metric's value.
	Value string
}

// ConverToMetric converts a PlainMetric to a strongly-typed Metric.
// It parses the string Value into the appropriate type based on the metric's Type.
// Returns an error if the Value cannot be parsed or if the Type is invalid.
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

// Metric represents a strongly-typed metric with appropriate value types.
// For gauge metrics, the Value field is used.
// For counter metrics, the Delta field is used.
type Metric struct {
	// Name is the identifier of the metric.
	Name string `json:"id"`
	// Value stores the value for gauge metrics (nil for counter metrics).
	Value *float64 `json:"value,omitempty"`
	// Delta stores the value for counter metrics (nil for gauge metrics).
	Delta *int64 `json:"delta,omitempty"`
	// Type specifies whether this is a gauge or counter metric.
	Type MetricType `json:"type"`
	// mapName is a cached string combining Type and Name for efficient lookups.
	mapName string
}

// String returns a string representation of the metric's value.
// For gauge metrics, it returns the float value formatted with up to 7 decimal places, with trailing zeros removed.
// For counter metrics, it returns the integer value as a string.
// For unknown metric types, it returns "unknown metric type".
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

// MapName returns a unique identifier string for the metric by combining its Type and Name.
// The format is "type-name" (e.g., "gauge-cpu" or "counter-requests").
// The result is cached in the mapName field for efficiency on subsequent calls.
func (m *Metric) MapName() string {
	if m.mapName == "" {
		m.mapName = string(m.Type) + "-" + m.Name
	}
	return m.mapName
}

// ConvertToPlain converts a strongly-typed Metric to a PlainMetric with string values.
// This is useful for serialization or when string representation is needed.
// The Value field of the returned PlainMetric is set using the String() method.
func (m *Metric) ConvertToPlain() *PlainMetric {
	return &PlainMetric{
		Name:  m.Name,
		Type:  m.Type,
		Value: m.String(),
	}
}

// NewMetric creates a new Metric from basic string parameters.
// It first creates a PlainMetric with the provided values, then converts it to a strongly-typed Metric.
// Returns an error if the conversion fails (e.g., if the value cannot be parsed or the type is invalid).
func NewMetric(metricType MetricType, metricName, metricValue string) (*Metric, error) {
	plainMetric := &PlainMetric{
		Name:  metricName,
		Type:  metricType,
		Value: metricValue,
	}
	return plainMetric.ConverToMetric()
}

package models

import (
	"errors"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"reflect"
	"testing"
)

func TestConvertToMetric(t *testing.T) {
	tests := []struct {
		name          string
		plainMetric   *PlainMetric
		expected      *Metric
		expectedError error
	}{
		{
			name: "Valid Gauge Metric",
			plainMetric: &PlainMetric{
				Name:  "temperature",
				Type:  GaugeType,
				Value: "25.5",
			},
			expected: &Metric{
				Name:  "temperature",
				Type:  GaugeType,
				Value: func() *float64 { v := 25.5; return &v }(),
			},
			expectedError: nil,
		},
		{
			name: "Valid Counter Metric",
			plainMetric: &PlainMetric{
				Name:  "requests",
				Type:  CounterType,
				Value: "42",
			},
			expected: &Metric{
				Name:  "requests",
				Type:  CounterType,
				Delta: func() *int64 { v := int64(42); return &v }(),
			},
			expectedError: nil,
		},
		{
			name: "Invalid Gauge Value",
			plainMetric: &PlainMetric{
				Name:  "temperature",
				Type:  GaugeType,
				Value: "invalid",
			},
			expected:      nil,
			expectedError: errors.New(errmsg.UnableToParseFloat),
		},
		{
			name: "Invalid Counter Value",
			plainMetric: &PlainMetric{
				Name:  "requests",
				Type:  CounterType,
				Value: "invalid",
			},
			expected:      nil,
			expectedError: errors.New(errmsg.UnableToParseInt),
		},
		{
			name: "Invalid Metric Type",
			plainMetric: &PlainMetric{
				Name:  "unknown",
				Type:  "invalid",
				Value: "100",
			},
			expected:      nil,
			expectedError: errors.New(errmsg.InvalidMetricType),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.plainMetric.ConverToMetric()

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected metric %v, got %v", tt.expected, result)
			}
		})
	}
}

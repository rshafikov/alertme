package metrics

import (
	"testing"

	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/stretchr/testify/assert"
)

func TestUpdateDataCollector(t *testing.T) {
	dc := NewEmptyDataCollector()

	t.Run("check PollCounter increments", func(t *testing.T) {
		initialValue := int64(0)
		dc.UpdateRuntimeMetrics()
		assert.Greater(t, *dc.PollCount.Delta, initialValue)
	})

	t.Run("check RandomValue changes", func(t *testing.T) {
		getMetricByName := func(name string, arr []*models.Metric) *models.Metric {
			for _, metric := range dc.Metrics {
				if metric.Name == name {
					return metric
				}
			}
			return nil
		}
		randValueBefore := getMetricByName("RandomValue", dc.Metrics).Value
		dc.UpdateRuntimeMetrics()
		randValueAfter := getMetricByName("RandomValue", dc.Metrics).Value

		assert.NotEqualValues(t, randValueBefore, randValueAfter)
	})
}

func BenchmarkDataCollector_CollectMetrics(b *testing.B) {
	dc := NewEmptyDataCollector()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		dc.UpdateRuntimeMetrics()
		dc.UpdateRuntimeMetrics()
	}
}

func TestUpdatePSUtilMetrics(t *testing.T) {
	dc := NewEmptyDataCollector()
	dc.UpdatePSUtilMetrics()
	
	if dc.TotalMemory == nil {
		t.Error("Expected TotalMemory to be initialized")
	}
	
	if dc.FreeMemory == nil {
		t.Error("Expected FreeMemory to be initialized")
	}
	
	if dc.CPUUtilization == nil {
		t.Error("Expected CPUUtilization to be initialized")
	}
}

func TestPassMetrics(t *testing.T) {
	dc := NewEmptyDataCollector()
	dc.UpdateRuntimeMetrics()
	dc.UpdatePSUtilMetrics()
	
	ch := make(chan []*models.Metric, 2)
	
	dc.PassMetrics(RuntimeMetrics, ch)
	
	select {
	case metrics := <-ch:
		if len(metrics) == 0 {
			t.Error("Expected non-empty metrics for RuntimeMetrics")
		}
	default:
		t.Error("Expected metrics to be sent to channel")
	}
	
	dc.PassMetrics(PSUtilMetrics, ch)
	
	select {
	case metrics := <-ch:
		_ = metrics
	default:
		t.Error("Expected metrics to be sent to channel")
	}
}

func TestNewEmptyDataCollector(t *testing.T) {
	dc := NewEmptyDataCollector()
	
	if dc.PollCount == nil {
		t.Error("Expected PollCount to be initialized")
	}
	
	if dc.PollCount.Name != "PollCount" {
		t.Errorf("Expected PollCount name to be 'PollCount', got '%s'", dc.PollCount.Name)
	}
	
	if dc.PollCount.Type != models.CounterType {
		t.Errorf("Expected PollCount type to be CounterType, got '%s'", dc.PollCount.Type)
	}
	
	if dc.PollCount.Delta == nil {
		t.Error("Expected PollCount.Delta to be initialized")
	}
	
	if *dc.PollCount.Delta != 0 {
		t.Errorf("Expected PollCount.Delta to be 0, got %d", *dc.PollCount.Delta)
	}
}

func TestFloat64Ptr(t *testing.T) {
	value := 1.23
	ptr := float64Ptr(value)
	
	if ptr == nil {
		t.Error("Expected pointer to be non-nil")
	}
	
	if ptr != nil && *ptr != value {
		t.Errorf("Expected value %f, got %f", value, *ptr)
	}
}

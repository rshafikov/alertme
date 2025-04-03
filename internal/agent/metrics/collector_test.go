package metrics

import (
	"github.com/rshafikov/alertme/internal/server/models"
	"testing"

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

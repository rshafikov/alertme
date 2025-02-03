package agent

import (
	"github.com/rshafikov/alertme/internal/server/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateDataCollector(t *testing.T) {
	dc := NewEmptyDataCollector()

	t.Run("check PollCounter increments", func(t *testing.T) {
		initialValue := dc.PollCount.Value
		UpdateDataCollector(dc)
		assert.Greater(t, dc.PollCount.Value, initialValue)
	})

	t.Run("check RandomValue changes", func(t *testing.T) {
		getMetricByName := func(name string, arr *[]models.GaugeMetric) models.GaugeMetric {
			for _, metric := range dc.Metrics {
				if metric.Name == name {
					return metric
				}
			}
			return models.GaugeMetric{}
		}
		randValueBefore := getMetricByName("RandomValue", &dc.Metrics).Value
		UpdateDataCollector(dc)
		randValueAfter := getMetricByName("RandomValue", &dc.Metrics).Value

		assert.NotEqualValues(t, randValueBefore, randValueAfter)
	})
}

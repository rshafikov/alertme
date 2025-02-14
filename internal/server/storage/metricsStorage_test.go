package storage

import (
	"testing"

	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_AddAndGet(t *testing.T) {
	storage := NewMemStorage()

	t.Run("add and get gauge metric", func(t *testing.T) {
		metric := &models.Metric{
			Type:  models.GaugeType,
			Name:  "myGauge",
			Value: "7777.777",
		}
		err := storage.Add(metric)
		require.NoError(t, err)

		got, err := storage.Get(models.GaugeType, "myGauge")
		require.NoError(t, err)
		assert.Equal(t, metric.Name, got.Name)
		assert.Equal(t, metric.Type, got.Type)
		assert.Equal(t, "7777.777", got.Value)
	})

	t.Run("add and get counter metric", func(t *testing.T) {
		metric := &models.Metric{
			Type:  models.CounterType,
			Name:  "myCounter",
			Value: "10",
		}
		err := storage.Add(metric)
		require.NoError(t, err)

		got, err := storage.Get(models.CounterType, "myCounter")
		require.NoError(t, err)
		assert.Equal(t, metric.Name, got.Name)
		assert.Equal(t, metric.Type, got.Type)
		assert.Equal(t, "10", got.Value)

		metric.Value = "5"
		err = storage.Add(metric)
		require.NoError(t, err)

		got, err = storage.Get(models.CounterType, "myCounter")
		require.NoError(t, err)
		assert.Equal(t, "15", got.Value) // 10 + 5
	})

	t.Run("add unsupported metric type", func(t *testing.T) {
		metric := &models.Metric{
			Type:  "unknown",
			Name:  "test_metric",
			Value: "100",
		}
		err := storage.Add(metric)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), UnsupportedMetricTypeErrMsg)
	})

	t.Run("add invalid gauge value", func(t *testing.T) {
		metric := &models.Metric{
			Type:  models.GaugeType,
			Name:  "invalid_gauge",
			Value: "not_a_number",
		}
		err := storage.Add(metric)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), CannotConvertToFloatErrMsg)
	})

	t.Run("add invalid counter value", func(t *testing.T) {
		metric := &models.Metric{
			Type:  models.CounterType,
			Name:  "invalid_counter",
			Value: "not_an_int",
		}
		err := storage.Add(metric)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), CannotConvertToIntErrMsg)
	})
}

func TestMemStorage_GetErrors(t *testing.T) {
	storage := NewMemStorage()

	t.Run("get metric which not exists", func(t *testing.T) {
		_, err := storage.Get(models.GaugeType, "missing_metric")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), NotFoundMetricErrMsg)
	})

	t.Run("get unsupported metric type", func(t *testing.T) {
		_, err := storage.Get("unknown", "test_metric")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), UnsupportedMetricTypeErrMsg)
	})
}

func TestMemStorage_Clear(t *testing.T) {
	storage := NewMemStorage()

	_ = storage.Add(&models.Metric{Type: models.GaugeType, Name: "CPU", Value: "60.0"})
	_ = storage.Add(&models.Metric{Type: models.CounterType, Name: "UPTIME", Value: "100"})

	storage.Clear()

	t.Run("check if storage is empty", func(t *testing.T) {
		_, err := storage.Get(models.GaugeType, "CPU")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), NotFoundMetricErrMsg)

		_, err = storage.Get(models.CounterType, "UPTIME")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), NotFoundMetricErrMsg)
	})
}

package storage

import (
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"testing"

	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_AddAndGet(t *testing.T) {
	storage := NewMemStorage()

	t.Run("add and get gauge metric", func(t *testing.T) {
		metric, err := models.NewMetric(models.GaugeType, "myGauge", "7777.777")
		require.NoError(t, err)

		err = storage.Add(metric)
		require.NoError(t, err)

		got, err := storage.Get(models.GaugeType, "myGauge")
		require.NoError(t, err)
		assert.Equal(t, metric.Name, got.Name)
		assert.Equal(t, metric.Type, got.Type)
		assert.EqualValues(t, 7777.777, *got.Value)
	})

	t.Run("add same gauge metric", func(t *testing.T) {
		metric, err := models.NewMetric(models.GaugeType, "myGauge", "7777.777")
		require.NoError(t, err)

		err = storage.Add(metric)
		require.NoError(t, err)

		got, err := storage.Get(models.GaugeType, "myGauge")
		require.NoError(t, err)
		assert.Equal(t, metric.Name, got.Name)
		assert.Equal(t, metric.Type, got.Type)
		assert.EqualValues(t, 7777.777, *got.Value)
	})

	t.Run("add and get counter metric", func(t *testing.T) {
		metric, err := models.NewMetric(models.CounterType, "myCounter", "10")
		require.NoError(t, err)

		err = storage.Add(metric)
		require.NoError(t, err)

		got, err := storage.Get(models.CounterType, "myCounter")
		require.NoError(t, err)
		assert.Equal(t, metric.Name, got.Name)
		assert.Equal(t, metric.Type, got.Type)
		assert.EqualValues(t, 10, *got.Delta)

		newSameMetric, err := models.NewMetric(models.CounterType, "myCounter", "90")
		require.NoError(t, err)

		err = storage.Add(newSameMetric)
		require.NoError(t, err)

		got, err = storage.Get(models.CounterType, "myCounter")
		require.NoError(t, err)
		assert.EqualValues(t, 100, *got.Delta)
	})

	t.Run("add unsupported metric type", func(t *testing.T) {
		value := 100.0
		metric := &models.Metric{
			Type:  "unknown",
			Name:  "test_metric",
			Value: &value,
		}
		err := storage.Add(metric)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errmsg.InvalidMetricType)
	})
}

func TestMemStorage_GetErrors(t *testing.T) {
	storage := NewMemStorage()

	t.Run("get metric which not exists", func(t *testing.T) {
		_, err := storage.Get(models.GaugeType, "missing_metric")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errmsg.MetricNotFound)
	})

	t.Run("get unsupported metric type", func(t *testing.T) {
		_, err := storage.Get("unknown", "test_metric")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errmsg.MetricNotFound)
	})
}

func TestMemStorage_Clear(t *testing.T) {
	storage := NewMemStorage()

	var (
		value = 60.0
		delta = int64(100)
	)
	_ = storage.Add(&models.Metric{Type: models.GaugeType, Name: "CPU", Value: &value})
	_ = storage.Add(&models.Metric{Type: models.CounterType, Name: "UPTIME", Delta: &delta})

	storage.Clear()

	t.Run("check if storage is empty", func(t *testing.T) {
		_, err := storage.Get(models.GaugeType, "CPU")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errmsg.MetricNotFound)

		_, err = storage.Get(models.CounterType, "UPTIME")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errmsg.MetricNotFound)
	})
}

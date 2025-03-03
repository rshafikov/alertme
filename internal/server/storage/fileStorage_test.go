package storage

import (
	"encoding/json"
	"fmt"
	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

const (
	TestFileName = "metrics.test.txt"
)

func TestFileLoader_SaveAndLoad(t *testing.T) {
	counter, err := models.NewMetric(models.CounterType, "veryImportantCounter", "777")
	assert.NoError(t, err)
	gauge, err := models.NewMetric(models.GaugeType, "veryImportantGauge", "777.7777")
	assert.NoError(t, err)
	expectedMetrics := []*models.Metric{counter, gauge}

	t.Run("Save and load metrics from a file", func(t *testing.T) {
		fileLoader := NewFileSaver(nil, TestFileName)
		saveErr := fileLoader.SaveMetrics([]*models.Metric{counter, gauge})
		assert.NoError(t, saveErr)

		metrics, loadErr := fileLoader.LoadMetrics()
		assert.NoError(t, loadErr)
		assert.Equal(t, expectedMetrics, metrics)
	})

	t.Run("Check if file exists", func(t *testing.T) {
		defer os.Remove(TestFileName)

		fileLoader := NewFileSaver(nil, TestFileName)
		_, checkErr := os.Stat(fileLoader.FileName)
		assert.NoError(t, checkErr)
		defer os.Remove(fileLoader.FileName)
	})
}

func TestFileLoader_LoadAndSaveToStorage(t *testing.T) {
	counter, err := models.NewMetric(models.CounterType, "CounterButNamesGauge", "777")
	assert.NoError(t, err)
	gauge, err := models.NewMetric(models.GaugeType, "blablalba", "12345678.90")
	assert.NoError(t, err)
	expectedMetrics := []*models.Metric{counter, gauge}

	t.Run("Load file not found", func(t *testing.T) {
		expectedErrMsg := fmt.Sprintf("open %s: no such file or directory", TestFileName)
		storage := NewMemStorage()
		fileLoader := NewFileSaver(storage, TestFileName)
		loadErr := fileLoader.LoadStorage()
		assert.Error(t, loadErr)
		assert.EqualError(t, loadErr, expectedErrMsg)
	})

	t.Run("Save metrics from storage to a file", func(t *testing.T) {
		storage := NewMemStorage()
		for _, metric := range expectedMetrics {
			err = storage.Add(metric)
			require.NoError(t, err)
		}

		fileLoader := NewFileSaver(storage, TestFileName)
		err = fileLoader.SaveStorage()
		defer os.Remove(TestFileName)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		_, checkErr := os.Stat(TestFileName)
		assert.NoError(t, checkErr)
	})

	t.Run("Load metrics from a file", func(t *testing.T) {
		defer os.Remove(TestFileName)

		_, createErr := os.Create(TestFileName)
		assert.NoError(t, createErr)

		testMetric1, err := models.NewMetric(models.CounterType, "a", "228")
		assert.NoError(t, err)
		testMetric2, err := models.NewMetric(models.CounterType, "b", "1337")
		assert.NoError(t, err)

		metricsList := []*models.Metric{testMetric1, testMetric2}
		file, openErr := os.OpenFile(TestFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		assert.NoError(t, openErr)
		defer file.Close()

		for _, metric := range metricsList {
			encoder := json.NewEncoder(file)
			encodeErr := encoder.Encode(metric)
			assert.NoError(t, encodeErr)
		}

		storage := NewMemStorage()
		fileLoader := NewFileSaver(storage, TestFileName)

		loadMetricsErr := fileLoader.LoadStorage()
		assert.NoError(t, loadMetricsErr)

		for _, metric := range metricsList {
			storedMetric, getErr := storage.Get(metric.Type, metric.Name)
			assert.NoError(t, getErr)
			assert.Equal(t, metric, storedMetric)
		}
	})
}

package storage

import (
	"context"
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
		ctx := context.Background()

		expectedErrMsg := fmt.Sprintf("open %s: no such file or directory", TestFileName)
		storage := NewMemStorage()
		fileLoader := NewFileSaver(storage, TestFileName)
		loadErr := fileLoader.LoadStorage(ctx)
		assert.Error(t, loadErr)
		assert.EqualError(t, loadErr, expectedErrMsg)
	})

	t.Run("Save metrics from storage to a file", func(t *testing.T) {
		ctx := context.Background()

		storage := NewMemStorage()
		for _, metric := range expectedMetrics {
			err = storage.Add(ctx, metric)
			require.NoError(t, err)
		}

		fileLoader := NewFileSaver(storage, TestFileName)
		err = fileLoader.SaveStorage(ctx)
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

		ctx := context.Background()

		loadMetricsErr := fileLoader.LoadStorage(ctx)
		assert.NoError(t, loadMetricsErr)

		for _, metric := range metricsList {
			storedMetric, getErr := storage.Get(ctx, metric.Type, metric.Name)
			assert.NoError(t, getErr)
			assert.Equal(t, metric, storedMetric)
		}
	})
}

func BenchmarkFileSaver_SaveMetrics(b *testing.B) {
	tempFile, err := os.CreateTemp("", "benchmark-metrics-*.txt")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	tempFileName := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempFileName)

	benchCases := []struct {
		name        string
		metricCount int
	}{
		{"Small", 10},
		{"Medium", 100},
		{"Large", 1000},
		{"XLarge", 10000},
	}

	for _, bc := range benchCases {
		b.Run(fmt.Sprintf("%s-%d", bc.name, bc.metricCount), func(b *testing.B) {
			metrics := generateTestMetrics(bc.metricCount)

			fileSaver := NewFileSaver(nil, tempFileName)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				file, truncErr := os.OpenFile(tempFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
				if truncErr != nil {
					b.Fatalf("Failed to truncate file: %v", truncErr)
				}
				file.Close()

				err := fileSaver.SaveMetrics(metrics)
				if err != nil {
					b.Fatalf("SaveMetrics failed: %v", err)
				}
			}
		})
	}
}

func generateTestMetrics(count int) []*models.Metric {
	metrics := make([]*models.Metric, 0, count)

	for i := 0; i < count; i++ {
		if i%2 == 0 {
			value := float64(i) * 1.5
			metrics = append(metrics, &models.Metric{
				Name:  fmt.Sprintf("gauge_metric_%d", i),
				Type:  models.GaugeType,
				Value: &value,
			})
		} else {
			delta := int64(i)
			metrics = append(metrics, &models.Metric{
				Name:  fmt.Sprintf("counter_metric_%d", i),
				Type:  models.CounterType,
				Delta: &delta,
			})
		}
	}

	return metrics
}

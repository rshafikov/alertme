package metrics

import (
	"fmt"
	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/rshafikov/alertme/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

const indexPath = "/"

func TestMetricsHandler_ListMetrics(t *testing.T) {
	memStorage := storage.NewMemStorage()
	router := NewMetricsRouter(memStorage)
	ts := httptest.NewServer(router.Routes())
	defer ts.Close()

	var notCompress bool
	client := NewHTTPClient(ts.URL+indexPath, notCompress)

	testGaugeMetric1 := models.Metric{
		Value: "0.0000001",
		Name:  "gauge1",
		Type:  models.GaugeType,
	}
	testGaugeMetric2 := models.Metric{
		Value: "1232.0000002",
		Name:  "gauge2",
		Type:  models.GaugeType,
	}
	testCounterMetric1 := models.Metric{
		Value: "1",
		Name:  "counter1",
		Type:  models.CounterType,
	}
	testCounterMetric2 := models.Metric{
		Value: "12321321321312312",
		Name:  "counter2",
		Type:  models.CounterType,
	}
	metrics := []models.Metric{testGaugeMetric1, testGaugeMetric2, testCounterMetric1, testCounterMetric2}
	err := FillStorageWithTestData(memStorage, metrics)
	require.NoError(t, err)

	t.Run("get all metrics", func(t *testing.T) {
		resp, respBody := client.URLRequest(t, http.MethodGet, "")
		defer resp.Body.Close()
		fmt.Println(respBody)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
		assert.Contains(t, respBody, testGaugeMetric1.Name)
		assert.Contains(t, respBody, testGaugeMetric1.Value)
		assert.Contains(t, respBody, testGaugeMetric2.Name)
		assert.Contains(t, respBody, testGaugeMetric2.Value)
		assert.Contains(t, respBody, testCounterMetric1.Name)
		assert.Contains(t, respBody, testCounterMetric1.Value)
		assert.Contains(t, respBody, testCounterMetric2.Name)
		assert.Contains(t, respBody, testCounterMetric2.Value)
	})
}

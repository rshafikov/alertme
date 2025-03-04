package metrics

import (
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/rshafikov/alertme/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const getResourceRESTPath = "/value"
const getResourceRESTMethod = http.MethodGet

func TestMetricsHandler_GetMetric(t *testing.T) {
	memStorage := storage.NewMemStorage()
	router := NewMetricsRouter(memStorage)
	ts := httptest.NewServer(router.Routes())
	defer ts.Close()

	var notCompress bool
	client := NewHTTPClient(ts.URL+getResourceRESTPath, notCompress)

	testGaugeMetric1 := models.PlainMetric{
		Value: "0.0000001",
		Name:  "gauge1",
		Type:  models.GaugeType,
	}
	testGaugeMetric2 := models.PlainMetric{
		Value: "1232.0000002",
		Name:  "gauge2",
		Type:  models.GaugeType,
	}
	testCounterMetric1 := models.PlainMetric{
		Value: "1",
		Name:  "counter1",
		Type:  models.CounterType,
	}
	testCounterMetric2 := models.PlainMetric{
		Value: "12321321321312312",
		Name:  "counter2",
		Type:  models.CounterType,
	}
	metrics := []models.PlainMetric{testGaugeMetric1, testGaugeMetric2, testCounterMetric1, testCounterMetric2}
	err := FillStorageWithTestData(memStorage, metrics)
	require.NoError(t, err)

	tests := []struct {
		name                string
		url                 string
		expectedCode        int
		expectedResponse    string
		expectedContentType string
	}{
		{
			name:                "get a counter metric #1",
			url:                 "/" + string(testCounterMetric1.Type) + "/" + testCounterMetric1.Name,
			expectedCode:        http.StatusOK,
			expectedResponse:    testCounterMetric1.Value,
			expectedContentType: "text/plain; charset=utf-8",
		},
		{
			name:                "get a counter metric #2",
			url:                 "/" + string(testCounterMetric2.Type) + "/" + testCounterMetric2.Name,
			expectedCode:        http.StatusOK,
			expectedResponse:    testCounterMetric2.Value,
			expectedContentType: "text/plain; charset=utf-8",
		},
		{
			name:                "get a gauge metric #1",
			url:                 "/" + string(testGaugeMetric1.Type) + "/" + testGaugeMetric1.Name,
			expectedCode:        http.StatusOK,
			expectedResponse:    testGaugeMetric1.Value,
			expectedContentType: "text/plain; charset=utf-8",
		},
		{
			name:                "get a gauge metric #2",
			url:                 "/" + string(testGaugeMetric2.Type) + "/" + testGaugeMetric2.Name,
			expectedCode:        http.StatusOK,
			expectedResponse:    testGaugeMetric2.Value,
			expectedContentType: "text/plain; charset=utf-8",
		},
		{
			name:                "get a metric with unknown type",
			url:                 "/unknownType/someName",
			expectedCode:        http.StatusBadRequest,
			expectedResponse:    errmsg.InvalidMetricType,
			expectedContentType: "text/plain; charset=utf-8",
		},
		{
			name:                "get a metric with unknown name",
			url:                 "/counter/someName",
			expectedCode:        http.StatusNotFound,
			expectedResponse:    errmsg.MetricNotFound,
			expectedContentType: "text/plain; charset=utf-8",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, respBody := client.URLRequest(t, getResourceRESTMethod, test.url)
			defer resp.Body.Close()

			assert.Equal(t, test.expectedCode, resp.StatusCode)
			assert.Equal(t, test.expectedResponse, strings.TrimRight(respBody, "\n"))
			assert.Equal(t, test.expectedContentType, resp.Header.Get("Content-Type"))
		})
	}
}

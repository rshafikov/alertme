package metrics

import (
	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/rshafikov/alertme/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsHandler_GetMetric(t *testing.T) {
	const baseHandlerPath = "/value"
	const handlerMethod = http.MethodGet
	memStorage := storage.NewMemStorage()
	router := NewMetricsRouter(memStorage)
	ts := httptest.NewServer(router.Routes())
	defer ts.Close()

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

	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "get a counter metric #1",
			url:  "/" + string(testCounterMetric1.Type) + "/" + testCounterMetric1.Name,
			want: want{
				code:        http.StatusOK,
				response:    testCounterMetric1.Value,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "get a counter metric #2",
			url:  "/" + string(testCounterMetric2.Type) + "/" + testCounterMetric2.Name,
			want: want{
				code:        http.StatusOK,
				response:    testCounterMetric2.Value,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "get a gauge metric #1",
			url:  "/" + string(testGaugeMetric1.Type) + "/" + testGaugeMetric1.Name,
			want: want{
				code:        http.StatusOK,
				response:    testGaugeMetric1.Value,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "get a gauge metric #2",
			url:  "/" + string(testGaugeMetric2.Type) + "/" + testGaugeMetric2.Name,
			want: want{
				code:        http.StatusOK,
				response:    testGaugeMetric2.Value,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "get a metric with unknown type",
			url:  "/unknownType/someName",
			want: want{
				code:        http.StatusBadRequest,
				response:    "invalid metric type\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "get a metric with unknown name",
			url:  "/counter/someName",
			want: want{
				code:        http.StatusNotFound,
				response:    "cannot find metric in storage\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, respBody := testRequest(t, ts, handlerMethod, baseHandlerPath+test.url)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)
			assert.Equal(t, test.want.response, respBody)
			assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

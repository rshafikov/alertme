package routers

import (
	"fmt"
	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/rshafikov/alertme/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func FillStorageWithTestData(s *storage.MemStorage, metrics []models.Metric) error {
	for _, metric := range metrics {
		if err := s.Add(&metric); err != nil {
			return err
		}
	}
	return nil
}

func TestMetricsHandler_CreateMetric(t *testing.T) {
	const baseHandlerPath = "/update"
	const handlerMethod = http.MethodPost
	memStorage := storage.NewMemStorage()
	router := NewMetricsRouter(memStorage)
	ts := httptest.NewServer(router.Routes())
	defer ts.Close()

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
			name: "create a counter metric",
			url:  "/counter/counterName/123",
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "create the same counter metric",
			url:  "/counter/counterName/123",
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "create a gauge metric",
			url:  "/gauge/gaugeName/321",
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "create the same gauge metric",
			url:  "/gauge/gaugeName/321",
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "create a metric with wrong type",
			url:  "/unexistedmetrictype/someName/111",
			want: want{
				code:        http.StatusBadRequest,
				response:    "invalid metric type\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "create a metric without name",
			url:  "/gauge/111",
			want: want{
				code:        http.StatusNotFound,
				response:    "404 page not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "create a metric with empty name",
			url:  "/gauge//111",
			want: want{
				code:        http.StatusNotFound,
				response:    "metric name is required\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "create gauge metric with incorrect value",
			url:  "/gauge/myGauge/123a",
			want: want{
				code:        http.StatusBadRequest,
				response:    "invalid metric value\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "create counter metric with incorrect value",
			url:  "/counter/myCounter/123a",
			want: want{
				code:        http.StatusBadRequest,
				response:    "invalid metric value\n",
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

func TestMetricsHandler_ListMetrics(t *testing.T) {
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

	t.Run("get all metrics", func(t *testing.T) {
		resp, respBody := testRequest(t, ts, http.MethodGet, "/")
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

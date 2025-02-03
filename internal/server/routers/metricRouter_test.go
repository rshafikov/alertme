package routers

import (
	"github.com/rshafikov/alertme/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const DefaultMetricURL = "/gauge/metricName/123"

func TestMetricsHandler_ServeHTTP(t *testing.T) {
	memStorage := storage.NewMemStorage()
	router := NewMetricsRouter(memStorage)
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		method string
		want   want
	}{
		{
			name:   "on GET request",
			method: http.MethodGet,
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    "Method Not Allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "on PUT request",
			method: http.MethodPut,
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    "Method Not Allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "on DELETE request",
			method: http.MethodDelete,
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    "Method Not Allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "on POST request",
			method: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				response:    ``,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.method, DefaultMetricURL, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestMetricsHandler_CreateMetric(t *testing.T) {
	memStorage := storage.NewMemStorage()
	router := NewMetricsRouter(memStorage)
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
				response:    "Not Found\n",
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
			r := httptest.NewRequest(http.MethodPost, test.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

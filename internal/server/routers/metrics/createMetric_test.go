package metrics

import (
	"github.com/rshafikov/alertme/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

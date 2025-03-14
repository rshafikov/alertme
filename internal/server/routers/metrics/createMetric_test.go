package metrics

import (
	"bytes"
	"compress/gzip"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const createResourceRESTPath = "/update"
const createResourceRESTMethod = http.MethodPost

func TestMetricsHandler_CreatePlaneMetric(t *testing.T) {
	memStorage := storage.NewMemStorage()
	router := NewMetricsRouter(memStorage, nil)
	ts := httptest.NewServer(router.Routes())
	defer ts.Close()

	var notCompress bool
	client := NewHTTPClient(ts.URL+createResourceRESTPath, notCompress)

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
				response:    errmsg.InvalidMetricType,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "create a metric without name",
			url:  "/gauge/111",
			want: want{
				code:        http.StatusNotFound,
				response:    "404 page not found",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "create a metric with empty name",
			url:  "/gauge//111",
			want: want{
				code:        http.StatusNotFound,
				response:    errmsg.MetricNameRequired,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "create gauge metric with incorrect value",
			url:  "/gauge/myGauge/123a",
			want: want{
				code:        http.StatusBadRequest,
				response:    errmsg.UnableToParseFloat,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "create counter metric with incorrect value",
			url:  "/counter/myCounter/123a",
			want: want{
				code:        http.StatusBadRequest,
				response:    errmsg.UnableToParseInt,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, respBody := client.URLRequest(t, createResourceRESTMethod, test.url)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)
			assert.Equal(t, test.want.response, strings.Trim(respBody, "\n"))
			assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

func TestMetricsHandler_CreateJSONMetric(t *testing.T) {
	memStorage := storage.NewMemStorage()
	router := NewMetricsRouter(memStorage, nil)
	ts := httptest.NewServer(router.Routes())
	defer ts.Close()

	var notCompress bool
	client := NewHTTPClient(ts.URL, notCompress)

	tests := []struct {
		name         string
		reqBody      string
		expectedCode int
		expectedBody string
		contentType  string
	}{
		{
			name:         "create a counter metric from JSON",
			expectedCode: http.StatusOK,
			reqBody:      `{"id": "counter_1", "delta": 1000, "type": "counter"}`,
			expectedBody: `{"id": "counter_1", "delta": 1000, "type": "counter"}`,
			contentType:  "application/json",
		},
		{
			name:         "create the same counter metric from JSON",
			expectedCode: http.StatusOK,
			reqBody:      `{"id": "counter_1", "delta": 1000, "type": "counter"}`,
			expectedBody: `{"id": "counter_1", "delta": 2000, "type": "counter"}`,
			contentType:  "application/json",
		},
		{
			name:         "create a gauge metric from JSON",
			expectedCode: http.StatusOK,
			reqBody:      `{"id": "gauge_1", "value": 123.4567, "type": "gauge"}`,
			expectedBody: `{"id": "gauge_1", "value": 123.4567, "type": "gauge"}`,
			contentType:  "application/json",
		},
		{
			name:         "create the same gauge metric from JSON",
			expectedCode: http.StatusOK,
			reqBody:      `{"id": "gauge_1", "value": 1234567, "type": "gauge"}`,
			expectedBody: `{"id": "gauge_1", "value": 1234567, "type": "gauge"}`,
			contentType:  "application/json",
		},
		{
			name:         "create a metric with wrong type from JSON",
			expectedCode: http.StatusBadRequest,
			reqBody:      `{"id": "gauge_1", "value": 1234567, "type": "gague"}`,
			expectedBody: errmsg.InvalidMetricType,
			contentType:  "text/plain",
		},
		{
			name:         "create a metric without name from JSON",
			expectedCode: http.StatusNotFound,
			reqBody:      `{"value": 1234567, "type": "gague"}`,
			expectedBody: errmsg.MetricNameRequired,
			contentType:  "text/plain",
		},
		{
			name:         "create a metric with empty name from JSON",
			expectedCode: http.StatusNotFound,
			reqBody:      `{"value": 1234567, "type": "gague"}`,
			expectedBody: errmsg.MetricNameRequired,
			contentType:  "text/plain",
		},
		{
			name:         "create gauge metric with incorrect value from JSON",
			expectedCode: http.StatusBadRequest,
			reqBody:      `{"id": "wrongVal", value": 12345d67, "type": "gague"}`,
			expectedBody: errmsg.UnableToDecodeJSON,
			contentType:  "text/plain",
		},
		{
			name:         "create counter metric with incorrect value from JSON",
			expectedCode: http.StatusBadRequest,
			reqBody:      `{"id": "wrongVal", delta": 12345d67, "type": "gague"}`,
			expectedBody: errmsg.UnableToDecodeJSON,
			contentType:  "text/plain",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, respBody := client.JSONRequest(t, createResourceRESTMethod, createResourceRESTPath+"/", test.reqBody)
			defer resp.Body.Close()

			assert.Equal(t, test.expectedCode, resp.StatusCode)
			if test.contentType == "application/json" {
				assert.JSONEq(t, test.expectedBody, respBody)
			} else {
				assert.Equal(t, test.expectedBody, strings.Trim(respBody, "\n"))
			}
			assert.Contains(t, resp.Header.Get("Content-Type"), test.contentType)
		})
	}
}

func TestMetricsRouter_GZIPCompression(t *testing.T) {
	memStorage := storage.NewMemStorage()
	router := NewMetricsRouter(memStorage, nil)
	ts := httptest.NewServer(router.Routes())
	defer ts.Close()

	requestBody := `{"id": "counter_1", "value": 1000, "type": "gauge"}`
	successBody := `{"id": "counter_1", "value": 1000, "type": "gauge"}`

	t.Run("send metric in gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest(createResourceRESTMethod, ts.URL+createResourceRESTPath+"/", buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "")
		r.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest(createResourceRESTMethod, ts.URL+createResourceRESTPath+"/", buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))
	})
}

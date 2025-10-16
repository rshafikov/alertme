package metrics

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/rshafikov/alertme/internal/server/settings"
	"github.com/rshafikov/alertme/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const updateURLPath = "/update"

func TestMetricsHandler_CreatePlaneMetric(t *testing.T) {
	memStorage := storage.NewMemStorage()
	router := NewMetricsRouter(memStorage)
	ts := httptest.NewServer(router.Routes())
	defer ts.Close()

	var notCompress bool
	client := NewHTTPClient(ts.URL+updateURLPath, notCompress)

	type want struct {
		response    string
		contentType string
		code        int
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
			resp, respBody := client.URLRequest(t, http.MethodPost, test.url)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)
			assert.Equal(t, test.want.response, strings.Trim(respBody, "\n"))
			assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

func TestMetricsHandler_CreateJSONMetric(t *testing.T) {
	memStorage := storage.NewMemStorage()
	router := NewMetricsRouter(memStorage)
	ts := httptest.NewServer(router.Routes())
	defer ts.Close()

	var notCompress bool
	client := NewHTTPClient(ts.URL, notCompress)

	tests := []struct {
		name         string
		reqBody      string
		expectedBody string
		contentType  string
		expectedCode int
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
			resp, respBody := client.JSONRequest(t, http.MethodPost, updateURLPath+"/", test.reqBody)
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
	router := NewMetricsRouter(memStorage)
	ts := httptest.NewServer(router.Routes())
	defer ts.Close()

	successBody := `{"id": "test_gzipped_gauge_1", "value": 1000, "type": "gauge"}`

	t.Run("send metric in gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(successBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, ts.URL+"/update/", buf)
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

		_, err = memStorage.Get(context.Background(), models.GaugeType, "test_gzipped_gauge_1")
		require.NoError(t, err)
	})

	t.Run("accepts gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(successBody)
		r := httptest.NewRequest(http.MethodPost, ts.URL+"/update/", buf)
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

	t.Run("get gzipped metric", func(t *testing.T) {
		metricName := "gaugeMustBeGzipped"
		metricDeltaStr := "1234123123123123123"
		getRequestBody := `{"id": "gaugeMustBeGzipped", "type": "gauge"}`
		getSuccessBody := `{"id": "gaugeMustBeGzipped", "value": 1234123123123123123, "type": "gauge"}`

		m, err := models.NewMetric(models.GaugeType, metricName, metricDeltaStr)
		require.NoError(t, err)

		err = memStorage.Add(context.Background(), m)
		require.NoError(t, err)

		err = memStorage.Add(context.Background(), m)
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, ts.URL+"/value/", strings.NewReader(getRequestBody))
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err, "resp body")

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, getSuccessBody, string(b))
	})

}

func TestMetricsRouter_HashMiddleware(t *testing.T) {
	memStorage := storage.NewMemStorage()
	router := NewMetricsRouter(memStorage)
	ts := httptest.NewServer(router.Routes())
	defer ts.Close()

	key := "I voted for Trump"
	settings.CONF.Key = key
	h := hmac.New(sha256.New, []byte(key))

	t.Run("send signed and zipped data", func(t *testing.T) {
		reqBody := `[{"id": "h_1", "value": 1234.56789, "type": "gauge"}, {"id": "h_2", "delta": 123456789, "type": "counter"}]`
		h.Write([]byte(reqBody))
		hash := h.Sum(nil)
		defer h.Reset()

		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(reqBody))
		require.NoError(t, err)
		require.NoError(t, zb.Close())

		r := httptest.NewRequest(http.MethodPost, ts.URL+"/updates/", buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "gzip")
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("HashSHA256", hex.EncodeToString(hash))

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("get signed and zipped data", func(t *testing.T) {
		gaugeMetricRequest := `{"id": "h_1", "type": "gauge"}`
		h.Write([]byte(`{"value":1234.56789,"id":"h_1","type":"gauge"}`))
		hash := h.Sum(nil)
		r := httptest.NewRequest(http.MethodPost, ts.URL+"/value/", strings.NewReader(gaugeMetricRequest))
		r.RequestURI = ""

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, hex.EncodeToString(hash), resp.Header.Get("Hashsha256"))
	})
}

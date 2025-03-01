package metrics

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/alertme/internal/server/config"
	"github.com/rshafikov/alertme/internal/server/models"
	"net/http"
	"strconv"
)

const (
	MetricNameRequiredErrMsg = "metric name is required"
	InvalidMetricTypeErrMsg  = "invalid metric type"
	InvalidMetricValueErrMsg = "invalid metric value"
	UnableAddMetricErrMsg    = "cannot add metric to storage"
	UnableToDecodeJSONErrMsg = "cannot decode JSON body"
	UnableToEncodeJSONErrMsg = "cannot encode JSON body"
	UnableToFindMetricErrMsg = "cannot find metric in storage"
	UnableToWriteRespErrMsg  = "cannot write response body"
)

type RequestError struct {
	Message string
	Code    int
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("%v <- %s", e.Code, e.Message)
}

func NewRequestError(msg string, code int) *RequestError {
	return &RequestError{
		Message: msg,
		Code:    code,
	}
}

func RespondWithError(w http.ResponseWriter, requestError *RequestError) {
	http.Error(w, requestError.Message, requestError.Code)
}

func (h *Router) ParseMetricFromURL(r *http.Request) (*models.Metric, *RequestError) {
	var m models.Metric
	m.Type = models.MetricType(chi.URLParam(r, "metricType"))
	m.Name = chi.URLParam(r, "metricName")
	m.Value = chi.URLParam(r, "metricValue")

	if m.Name == "" {
		config.Log.Debug(MetricNameRequiredErrMsg)
		return &m, NewRequestError(MetricNameRequiredErrMsg, http.StatusNotFound)
	}

	if !(m.Type == models.CounterType || m.Type == models.GaugeType) {
		config.Log.Debug(InvalidMetricTypeErrMsg)
		return &m, NewRequestError(InvalidMetricTypeErrMsg, http.StatusBadRequest)
	}

	if r.Method != http.MethodGet {
		if _, err := strconv.ParseFloat(m.Value, 64); err != nil {
			config.Log.Debug(InvalidMetricValueErrMsg)
			return &m, NewRequestError(InvalidMetricValueErrMsg, http.StatusBadRequest)
		}
	}

	return &m, nil
}

func (h *Router) ParseMetricFromJSON(r *http.Request) (*models.Metric, *RequestError) {
	var reqMetric models.MetricJSONReq
	var m models.Metric

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&reqMetric); err != nil {
		config.Log.Debug(UnableToDecodeJSONErrMsg)
		return &m, NewRequestError(UnableToDecodeJSONErrMsg, http.StatusBadRequest)
	}

	m.Name = reqMetric.ID
	if m.Name == "" {
		config.Log.Debug(MetricNameRequiredErrMsg)
		return &m, NewRequestError(MetricNameRequiredErrMsg, http.StatusNotFound)
	}

	m.Type = models.MetricType(reqMetric.MType)
	if !(m.Type == models.CounterType || m.Type == models.GaugeType) {
		config.Log.Debug(InvalidMetricTypeErrMsg)
		return &m, NewRequestError(InvalidMetricTypeErrMsg, http.StatusBadRequest)
	}
	if r.URL.Path == "/update/" {
		switch m.Type {
		case models.CounterType:
			if reqMetric.Delta != nil {
				m.Value = strconv.FormatInt(*reqMetric.Delta, 10)
			} else {
				config.Log.Debug(InvalidMetricValueErrMsg)
				return &m, NewRequestError(InvalidMetricValueErrMsg, http.StatusBadRequest)
			}
		case models.GaugeType:
			if reqMetric.Value != nil {
				m.Value = strconv.FormatFloat(*reqMetric.Value, 'f', -1, 64)
			} else {
				config.Log.Debug(InvalidMetricValueErrMsg)
				return &m, NewRequestError(InvalidMetricValueErrMsg, http.StatusBadRequest)
			}
		}
	}

	return &m, nil
}

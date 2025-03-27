package metrics

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/models"
	"net/http"
)

func (h *Router) ParseMetricFromURL(r *http.Request) (*models.Metric, int, error) {
	metricType := models.MetricType(chi.URLParam(r, "metricType"))
	metricName := chi.URLParam(r, "metricName")
	metricStrValue := chi.URLParam(r, "metricValue")

	errCode, err := h.baseMetricValidation(metricName, metricType)
	if err != nil {
		return nil, errCode, err
	}

	if r.Method != http.MethodGet {
		newMetric, err := models.NewMetric(metricType, metricName, metricStrValue)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		return newMetric, http.StatusOK, nil
	}

	return &models.Metric{
		Name:  metricName,
		Value: nil,
		Delta: nil,
		Type:  metricType,
	}, http.StatusOK, nil
}

func (h *Router) ParseMetricFromJSON(r *http.Request) (*models.Metric, int, error) {
	var reqMetric models.Metric

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&reqMetric); err != nil {
		return nil, http.StatusBadRequest, errors.New(errmsg.UnableToDecodeJSON)
	}

	errCode, err := h.baseMetricValidation(reqMetric.Name, reqMetric.Type)
	if err != nil {
		return nil, errCode, err
	}

	if r.URL.Path == "/update/" {
		switch reqMetric.Type {
		case models.CounterType:
			if reqMetric.Delta == nil {
				return nil, http.StatusBadRequest, errors.New(errmsg.InvalidMetricValue)
			}
		case models.GaugeType:
			if reqMetric.Value == nil {
				return nil, http.StatusBadRequest, errors.New(errmsg.InvalidMetricValue)
			}
		}
	}

	return &reqMetric, http.StatusOK, nil
}

func (h *Router) baseMetricValidation(metricName string, metricType models.MetricType) (int, error) {
	if metricName == "" {
		return http.StatusNotFound, errors.New(errmsg.MetricNameRequired)
	}

	if !(metricType == models.CounterType || metricType == models.GaugeType) {
		return http.StatusBadRequest, errors.New(errmsg.InvalidMetricType)
	}

	return http.StatusOK, nil
}

func (h *Router) ParseMetricsFromJSON(r *http.Request) ([]*models.Metric, int, error) {
	var reqMetrics []*models.Metric

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&reqMetrics); err != nil {
		return nil, http.StatusBadRequest, errors.New(errmsg.UnableToDecodeJSON)
	}
	for _, reqMetric := range reqMetrics {
		errCode, err := h.baseMetricValidation(reqMetric.Name, reqMetric.Type)
		if err != nil {
			return nil, errCode, err
		}
	}
	return reqMetrics, http.StatusOK, nil
}

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

	if metricName == "" {
		return nil, http.StatusNotFound, errors.New(errmsg.MetricNameRequired)
	}

	if !(metricType == models.CounterType || metricType == models.GaugeType) {
		return nil, http.StatusBadRequest, errors.New(errmsg.InvalidMetricType)
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

	if reqMetric.Name == "" {
		return nil, http.StatusNotFound, errors.New(errmsg.MetricNameRequired)
	}

	if !(reqMetric.Type == models.CounterType || reqMetric.Type == models.GaugeType) {
		return nil, http.StatusBadRequest, errors.New(errmsg.InvalidMetricType)
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

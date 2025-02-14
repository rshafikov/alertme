package metrics

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/rshafikov/alertme/internal/server/storage"
	"log"
	"net/http"
	"strconv"
)

func (h *Router) CreateMetric(w http.ResponseWriter, r *http.Request) {
	metricType := models.MetricType(chi.URLParam(r, "metricType"))
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	if metricName == "" {
		log.Println("metric name is required")
		http.Error(w, "metric name is required", http.StatusNotFound)
		return
	}

	if !(metricType == models.CounterType || metricType == models.GaugeType) {
		log.Println("invalid metric type")
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	if _, err := strconv.ParseFloat(metricValue, 64); err != nil {
		log.Println("invalid metric value")
		http.Error(w, "invalid metric value", http.StatusBadRequest)
		return
	}

	newMetric := models.Metric{Type: metricType, Name: metricName, Value: metricValue}

	if err := h.store.Add(&newMetric); err != nil {
		var (
			metricTypeErr  *storage.UnsupportedMetricTypeError
			metricValueErr *storage.IncorrectMetricValueError
		)
		if errors.As(err, &metricTypeErr) {
			log.Println("ERR", metricTypeErr.Error())
			http.Error(w, "", http.StatusBadRequest)
			return

		} else if errors.As(err, &metricValueErr) {
			log.Println("ERR", metricValueErr.Error())
			http.Error(w, "", http.StatusBadRequest)
			return

		} else {
			log.Println("ERR: cannot add metric to storage", err)
			http.Error(w, "cannot add metric to storage", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

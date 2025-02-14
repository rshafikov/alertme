package metrics

import (
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/alertme/internal/server/models"
	"log"
	"net/http"
)

func (h *Router) GetMetric(w http.ResponseWriter, r *http.Request) {
	metricType := models.MetricType(chi.URLParam(r, "metricType"))
	metricName := chi.URLParam(r, "metricName")

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

	metric, err := h.store.Get(metricType, metricName)
	if err != nil {
		log.Println("ERR", "cannot find metric in storage", err)
		http.Error(w, "cannot find metric in storage", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(metric.Value))
}

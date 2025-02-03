package routers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/rshafikov/alertme/internal/server/storage"
)

type MetricsRouter struct {
	store storage.BaseMetricStorage
}

func NewMetricsRouter(s storage.BaseMetricStorage) *MetricsRouter {
	return &MetricsRouter{
		store: s,
	}
}

func (h *MetricsRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost:
		h.CreateMetric(w, r)
		return
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

}

func (h *MetricsRouter) CreateMetric(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	log.Println(params)
	if len(params) != 3 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	metricType, metricName, metricValue := models.MetricType(params[0]), params[1], params[2]

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

	log.Println("LOG:", "trying to create new metric:", newMetric)
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

	log.Println("LOG:", "metric was created")

	_, err := h.store.Get(newMetric.Type, newMetric.Name)
	if err != nil {
		log.Println("ERR", "cannot add metric to storage", err)
		http.Error(w, "cannot add metric to storage", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

}

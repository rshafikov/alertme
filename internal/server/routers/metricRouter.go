package routers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/rshafikov/alertme/internal/server/services"
)

type MetricsHandler struct {
	store services.BaseMetricStorage
}

func NewMetricsHandler(s services.BaseMetricStorage) *MetricsHandler {
	return &MetricsHandler{
		store: s,
	}
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost:
		h.CreateMetric(w, r)
		return
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

}

func (h *MetricsHandler) CreateMetric(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	fmt.Println(params)
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

	fmt.Println("LOG:", "trying to create new metric:", newMetric)
	if err := h.store.Add(&newMetric); err != nil {
		var (
			metricTypeErr  *services.UnsupportedMetricTypeError
			metricValueErr *services.IncorrectMetricValueError
		)
		if errors.As(err, &metricTypeErr) {
			fmt.Println("ERR", metricTypeErr.Error())
			http.Error(w, "", http.StatusBadRequest)
			return

		} else if errors.As(err, &metricValueErr) {
			fmt.Println("ERR", metricValueErr.Error())
			http.Error(w, "", http.StatusBadRequest)
			return

		} else {
			fmt.Println("ERR: cannot add metric to storage", err)
			http.Error(w, "cannot add metric to storage", http.StatusInternalServerError)
			return
		}
	}

	fmt.Println("LOG:", "metric was created")

	createdMetric, err := h.store.Get(newMetric.Type, newMetric.Name)
	if err != nil {
		fmt.Println("ERR", "cannot add metric to storage", err)
		http.Error(w, "cannot add metric to storage", http.StatusInternalServerError)
		return
	}

	jsonBytes, _ := json.Marshal(createdMetric)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonBytes)
	if err != nil {
		return
	}
}

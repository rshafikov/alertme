package routers

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/rshafikov/alertme/internal/server/storage"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type MetricsRouter struct {
	store storage.BaseMetricStorage
}

func NewMetricsRouter(store storage.BaseMetricStorage) *MetricsRouter {
	return &MetricsRouter{store: store}
}

func (h *MetricsRouter) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", h.ListMetrics)
	r.Route("/update", func(r chi.Router) {
		r.Post("/{metricType}/{metricName}/{metricValue}", h.CreateMetric)
	})
	r.Route("/value", func(r chi.Router) {
		r.Get("/{metricType}/{metricName}", h.GetMetric)
	})
	return r
}

func (h *MetricsRouter) CreateMetric(w http.ResponseWriter, r *http.Request) {
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

func (h *MetricsRouter) GetMetric(w http.ResponseWriter, r *http.Request) {
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

func (h *MetricsRouter) ListMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.store.List()
	if err != nil {
		log.Println("ERR", "cannot list metrics in storage", err)
		http.Error(w, "cannot list metrics in storage", http.StatusInternalServerError)
		return
	}

	const tmpl = `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Metrics</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 20px; }
			table { border-collapse: collapse; width: 50%; }
			th, td { border: 1px solid black; padding: 8px; text-align: left; }
			th { background-color: #f2f2f2; }
		</style>
	</head>
	<body>
		<h1>Metrics List</h1>
		<table>
			<tr><th>Name</th><th>Value</th></tr>
			{{range .}}
				<tr><td>{{.Name}}</td><td>{{.Value}}</td></tr>
			{{end}}
		</table>
	</body>
	</html>`

	t, err := template.New("metrics").Parse(tmpl)
	if err != nil {
		log.Println("ERR", "cannot parse template", err)
		http.Error(w, "cannot parse template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	err = t.Execute(w, metrics)
	if err != nil {
		log.Println("ERR", "cannot execute template", err)
		http.Error(w, "cannot execute template", http.StatusInternalServerError)
	}
}

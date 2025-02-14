package metrics

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rshafikov/alertme/internal/server/storage"
)

type Router struct {
	store storage.BaseMetricStorage
}

func NewMetricsRouter(store storage.BaseMetricStorage) *Router {
	return &Router{store: store}
}

func (h *Router) Routes() chi.Router {
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

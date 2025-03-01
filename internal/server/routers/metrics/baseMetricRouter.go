package metrics

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rshafikov/alertme/internal/server/middlewares"
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
	r.Use(middlewares.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.GZipper)

	r.Get("/", h.ListMetrics)
	r.Route("/update", func(r chi.Router) {
		r.Post("/", h.CreateMetricFromJSON)
		r.Post("/{metricType}/{metricName}/{metricValue}", h.CreateMetricFromURL)
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", h.GetMericFromJSON)
		r.Get("/{metricType}/{metricName}", h.GetMetricFromURL)
	})
	return r
}

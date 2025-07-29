package metrics

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rshafikov/alertme/internal/server/middlewares"
	"github.com/rshafikov/alertme/internal/server/storage"
)

// Router manages HTTP routes and interactions with the metric storage system.
type Router struct {
	store storage.BaseMetricStorage
}

// NewMetricsRouter initializes a new Router with the provided metric storage.
func NewMetricsRouter(store storage.BaseMetricStorage) *Router {
	return &Router{
		store: store,
	}
}

// Routes initializes and configures the application's routes and middleware stack.
// Returns a chi.Router instance with all routes and middleware applied.
func (h *Router) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.GZipper)
	r.Use(middlewares.Hasher)

	r.Get("/", h.ListMetrics)
	r.Get("/ping", h.PingDB)
	r.Post("/updates/", h.CreateMetricsFromJSON)
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

package metrics

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rshafikov/alertme/internal/server/middlewares"
	"github.com/rshafikov/alertme/internal/server/storage"
)

type Router struct {
	store storage.BaseMetricStorage
	db    *storage.DBStorage
}

func NewMetricsRouter(store storage.BaseMetricStorage, db *storage.DBStorage) *Router {
	return &Router{
		store: store,
		db:    db,
	}
}

func (h *Router) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.GZipper)

	r.Get("/", h.ListMetrics)
	r.Get("/ping", h.PingDB)
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

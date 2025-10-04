package metrics

import (
	"github.com/rshafikov/alertme/internal/server/logger"
	"go.uber.org/zap"
	"net/http"
)

// CreateMetricsFromJSON handles HTTP requests to create metrics from JSON payloads.
// It processes the request body to parse and validate metrics data.
// Metrics are stored in batch mode to the backend using the provided context.
// Responds with appropriate HTTP status codes based on success or error cases.
func (h *Router) CreateMetricsFromJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	newMetrics, responseCode, parseErr := h.ParseMetricsFromJSON(r)
	if parseErr != nil {
		logger.Log.Debug(parseErr.Error())
		http.Error(w, parseErr.Error(), responseCode)
		return
	}

	if saveErr := h.store.AddBatch(ctx, newMetrics); saveErr != nil {
		logger.Log.Debug(saveErr.Error())
		http.Error(w, saveErr.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(""))
	if err != nil {
		logger.Log.Error("unable to process metrics", zap.Error(err))
	}
}

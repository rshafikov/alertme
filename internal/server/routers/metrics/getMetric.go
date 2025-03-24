package metrics

import (
	"encoding/json"
	"errors"
	"github.com/rshafikov/alertme/internal/server/database"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/logger"
	"go.uber.org/zap"
	"net/http"
)

func (h *Router) GetMetricFromURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	parsedMetric, responseCode, parseErr := h.ParseMetricFromURL(r)
	if parseErr != nil {
		logger.Log.Debug(parseErr.Error())
		http.Error(w, parseErr.Error(), responseCode)
		return
	}

	storedMetric, saveErr := h.store.Get(ctx, parsedMetric.Type, parsedMetric.Name)
	if saveErr != nil {
		logger.Log.Debug("an error happened during request", zap.Error(saveErr))
		if errors.Is(saveErr, database.ErrDB) || errors.Is(saveErr, database.ErrConnToDB) {
			http.Error(w, saveErr.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(w, saveErr.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, writeErr := w.Write([]byte(storedMetric.String()))
	if writeErr != nil {
		logger.Log.Debug(errmsg.UnableToWriteResponse)
		return
	}
}

func (h *Router) GetMericFromJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	newMetric, responseCode, parseErr := h.ParseMetricFromJSON(r)
	if parseErr != nil {
		logger.Log.Debug(parseErr.Error())
		http.Error(w, parseErr.Error(), responseCode)
		return
	}

	storedMetric, getErr := h.store.Get(ctx, newMetric.Type, newMetric.Name)
	if getErr != nil {
		logger.Log.Debug(getErr.Error())
		http.Error(w, getErr.Error(), http.StatusNotFound)
		return
	}

	jsonBytes, encodeErr := json.Marshal(storedMetric)
	if encodeErr != nil {
		logger.Log.Debug(errmsg.UnableToEncodeJSON)
		http.Error(w, encodeErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, writeErr := w.Write(jsonBytes)
	if writeErr != nil {
		logger.Log.Debug(errmsg.UnableToWriteResponse)
		return
	}
}

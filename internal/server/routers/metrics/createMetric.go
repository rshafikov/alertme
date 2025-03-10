package metrics

import (
	"encoding/json"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/logger"
	"go.uber.org/zap"
	"net/http"
)

func (h *Router) CreateMetricFromURL(w http.ResponseWriter, r *http.Request) {
	newMetric, responseCode, parseErr := h.ParseMetricFromURL(r)

	if parseErr != nil {
		logger.Log.Debug("Unable to parse metric", zap.Error(parseErr))
		http.Error(w, parseErr.Error(), responseCode)
		return
	}

	if storageErr := h.store.Add(newMetric); storageErr != nil {
		logger.Log.Debug(storageErr.Error())
		http.Error(w, storageErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (h *Router) CreateMetricFromJSON(w http.ResponseWriter, r *http.Request) {
	newMetric, responseCode, parseErr := h.ParseMetricFromJSON(r)
	if parseErr != nil {
		logger.Log.Debug(parseErr.Error())
		http.Error(w, parseErr.Error(), responseCode)
		return
	}

	if saveErr := h.store.Add(newMetric); saveErr != nil {
		logger.Log.Debug(saveErr.Error())
		http.Error(w, saveErr.Error(), http.StatusInternalServerError)
		return
	}

	createdMetric, getErr := h.store.Get(newMetric.Type, newMetric.Name)
	if getErr != nil {
		logger.Log.Debug(getErr.Error())
		http.Error(w, getErr.Error(), http.StatusInternalServerError)
		return
	}

	jsonBytes, encodeErr := json.Marshal(createdMetric)
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

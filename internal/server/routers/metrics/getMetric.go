package metrics

import (
	"encoding/json"
	"github.com/rshafikov/alertme/internal/server/config"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"net/http"
)

func (h *Router) GetMetricFromURL(w http.ResponseWriter, r *http.Request) {
	parsedMetric, responseCode, parseErr := h.ParseMetricFromURL(r)
	if parseErr != nil {
		config.Log.Debug(parseErr.Error())
		http.Error(w, parseErr.Error(), responseCode)
		return
	}

	storedMetric, saveErr := h.store.Get(parsedMetric.Type, parsedMetric.Name)
	if saveErr != nil {
		config.Log.Debug(saveErr.Error())
		http.Error(w, saveErr.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, writeErr := w.Write([]byte(storedMetric.String()))
	if writeErr != nil {
		config.Log.Debug(errmsg.UnableToWriteResponse)
		return
	}
}

func (h *Router) GetMericFromJSON(w http.ResponseWriter, r *http.Request) {
	newMetric, responseCode, parseErr := h.ParseMetricFromJSON(r)
	if parseErr != nil {
		config.Log.Debug(parseErr.Error())
		http.Error(w, parseErr.Error(), responseCode)
		return
	}

	storedMetric, getErr := h.store.Get(newMetric.Type, newMetric.Name)
	if getErr != nil {
		config.Log.Debug(getErr.Error())
		http.Error(w, getErr.Error(), http.StatusNotFound)
		return
	}

	jsonBytes, encodeErr := json.Marshal(storedMetric)
	if encodeErr != nil {
		config.Log.Debug(errmsg.UnableToEncodeJSON)
		http.Error(w, encodeErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, writeErr := w.Write(jsonBytes)
	if writeErr != nil {
		config.Log.Debug(errmsg.UnableToWriteResponse)
		return
	}
}

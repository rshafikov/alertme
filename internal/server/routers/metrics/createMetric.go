package metrics

import (
	"encoding/json"
	"github.com/rshafikov/alertme/internal/server/config"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"net/http"
)

func (h *Router) CreateMetricFromURL(w http.ResponseWriter, r *http.Request) {
	newMetric, responseCode, parseErr := h.ParseMetricFromURL(r)

	if parseErr != nil {
		config.Log.Debug(parseErr.Error())
		http.Error(w, parseErr.Error(), responseCode)
		return
	}

	if storageErr := h.store.Add(newMetric); storageErr != nil {
		config.Log.Debug(storageErr.Error())
		http.Error(w, storageErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (h *Router) CreateMetricFromJSON(w http.ResponseWriter, r *http.Request) {
	newMetric, responseCode, parseErr := h.ParseMetricFromJSON(r)
	if parseErr != nil {
		config.Log.Debug(parseErr.Error())
		http.Error(w, parseErr.Error(), responseCode)
		return
	}

	if saveErr := h.store.Add(newMetric); saveErr != nil {
		config.Log.Debug(saveErr.Error())
		http.Error(w, saveErr.Error(), http.StatusInternalServerError)
		return
	}

	createdMetric, getErr := h.store.Get(newMetric.Type, newMetric.Name)
	if getErr != nil {
		config.Log.Debug(getErr.Error())
		http.Error(w, getErr.Error(), http.StatusInternalServerError)
		return
	}

	jsonBytes, encodeErr := json.Marshal(createdMetric)
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

package metrics

import (
	"encoding/json"
	"github.com/rshafikov/alertme/internal/server/config"
	"github.com/rshafikov/alertme/internal/server/models"
	"net/http"
)

func (h *Router) GetMetricFromURL(w http.ResponseWriter, r *http.Request) {
	metric, reqErr := h.ParseMetricFromURL(r)
	if reqErr != nil {
		RespondWithError(w, reqErr)
		return
	}

	storedMetric, err := h.store.Get(metric.Type, metric.Name)
	if err != nil {
		config.Log.Debug(UnableToFindMetricErrMsg)
		RespondWithError(w, NewRequestError(UnableToFindMetricErrMsg, http.StatusNotFound))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(storedMetric.Value))
	if err != nil {
		config.Log.Debug(UnableToWriteRespErrMsg)
		return
	}

}

func (h *Router) GetMericFromJSON(w http.ResponseWriter, r *http.Request) {
	metric, reqErr := h.ParseMetricFromJSON(r)
	if reqErr != nil {
		RespondWithError(w, reqErr)
		return
	}

	storedMetric, err := h.store.Get(metric.Type, metric.Name)
	if err != nil {
		config.Log.Debug(UnableToFindMetricErrMsg)
		RespondWithError(w, NewRequestError(UnableToFindMetricErrMsg, http.StatusNotFound))
		return
	}

	var respMetric models.MetricJSONReq
	respMetric.ConvertToBaseMetric(&storedMetric)

	jsonBytes, err := json.Marshal(respMetric)
	if err != nil {
		config.Log.Debug(UnableToEncodeJSONErrMsg)
		RespondWithError(w, NewRequestError(UnableToEncodeJSONErrMsg, http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonBytes)
	if err != nil {
		config.Log.Debug(UnableToWriteRespErrMsg)
		return
	}

}

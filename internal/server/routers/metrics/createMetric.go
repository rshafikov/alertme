package metrics

import (
	"encoding/json"
	"errors"
	"github.com/rshafikov/alertme/internal/server/config"
	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/rshafikov/alertme/internal/server/storage"
	"net/http"
)

func (h *Router) addMetricToStorage(metric *models.Metric) *RequestError {
	var (
		metricTypeErr  *storage.UnsupportedMetricTypeError
		metricValueErr *storage.IncorrectMetricValueError
	)

	err := h.store.Add(metric)
	if err != nil {
		if errors.As(err, &metricTypeErr) {
			config.Log.Debug(metricTypeErr.Error())
			return NewRequestError(UnableAddMetricErrMsg, http.StatusBadRequest)

		} else if errors.As(err, &metricValueErr) {
			config.Log.Debug(metricValueErr.Error())
			return NewRequestError(UnableAddMetricErrMsg, http.StatusBadRequest)

		} else {
			config.Log.Debug(UnableAddMetricErrMsg, err)
			return NewRequestError(UnableAddMetricErrMsg, http.StatusInternalServerError)
		}
	}
	return nil
}

func (h *Router) CreateMetricFromURL(w http.ResponseWriter, r *http.Request) {
	newMetric, parseErr := h.ParseMetricFromURL(r)

	if parseErr != nil {
		RespondWithError(w, parseErr)
		return
	}

	if storageErr := h.addMetricToStorage(newMetric); storageErr != nil {
		RespondWithError(w, storageErr)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (h *Router) CreateMetricFromJSON(w http.ResponseWriter, r *http.Request) {
	newMetric, reqErr := h.ParseMetricFromJSON(r)
	if reqErr != nil {
		RespondWithError(w, reqErr)
		return
	}
	if storageErr := h.addMetricToStorage(newMetric); storageErr != nil {
		RespondWithError(w, storageErr)
		return
	}
	createdBaseMetric, err := h.store.Get(newMetric.Type, newMetric.Name)
	if err != nil {
		config.Log.Debug(UnableAddMetricErrMsg)
		RespondWithError(w, NewRequestError(UnableAddMetricErrMsg, http.StatusInternalServerError))
		return
	}

	var respMetric models.MetricJSONReq
	respMetric.ConvertToBaseMetric(&createdBaseMetric)

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

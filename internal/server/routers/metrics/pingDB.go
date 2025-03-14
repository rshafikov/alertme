package metrics

import (
	"context"
	"errors"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/logger"
	"net/http"
)

func (h *Router) PingDB(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		logger.Log.Error(errmsg.UnableToPingDB)
		http.Error(w, errors.New(errmsg.UnableToPingDB).Error(), http.StatusInternalServerError)
		return
	}

	err := h.db.Pool.Ping(context.Background())
	if err != nil {
		logger.Log.Error(errmsg.UnableToPingDB)
		http.Error(w, errors.New(errmsg.UnableToPingDB).Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)

}

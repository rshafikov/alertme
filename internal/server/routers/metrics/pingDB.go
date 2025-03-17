package metrics

import (
	"errors"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/storage"
	"net/http"
)

func (h *Router) PingDB(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	db, ok := h.store.(storage.BaseDatabase)
	if !ok {
		logger.Log.Error(errmsg.UnableToPingDB)
		http.Error(w, errors.New(errmsg.UnableToPingDB).Error(), http.StatusInternalServerError)
		return
	}

	err := db.Ping(ctx)
	if err != nil {
		logger.Log.Error(errmsg.UnableToPingDB)
		http.Error(w, errors.New(errmsg.UnableToPingDB).Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)

}

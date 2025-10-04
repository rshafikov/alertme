package metrics

import (
	"errors"
	"github.com/rshafikov/alertme/internal/server/database"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/logger"
	"net/http"
)

// PingDB handles the request to check the connectivity status of the database.
// It responds with an HTTP status code 200 if successful, or 500 on failure.
func (h *Router) PingDB(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	db, ok := h.store.(database.Pinger)
	if !ok {
		logger.Log.Error(errmsg.UnableToPingDB)
		http.Error(w, errors.New(errmsg.UnableToPingDB).Error(), http.StatusInternalServerError)
		return
	}

	err := db.Ping(ctx)
	if err != nil {
		logger.Log.Error(errmsg.UnableToPingDB)
		http.Error(w, errors.New(errmsg.UnableToPingDB).Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

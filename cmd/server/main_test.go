package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rshafikov/alertme/internal/server/routers"
	"github.com/rshafikov/alertme/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

func Test_Run(t *testing.T) {
	store := storage.NewMemStorage()
	metricsRouter := routers.NewMetricsRouter(store)

	mux := http.NewServeMux()
	mux.Handle("/update/", http.StripPrefix("/update", metricsRouter))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	t.Run("check if POST /update/counter/myCounter/1 works", func(t *testing.T) {
		resp, err := http.Post(ts.URL+"/update/counter/test/1", "application/json", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		defer resp.Body.Close()
	})

	t.Run("check if GET /update/counter/test fails", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/update/counter/test")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		defer resp.Body.Close()
	})
}

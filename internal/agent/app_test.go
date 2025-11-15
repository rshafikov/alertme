package agent

import (
	"github.com/rshafikov/alertme/internal/agent/metrics"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestNewAgentApp(t *testing.T) {
	client := &Client{}
	dc := &metrics.DataCollector{}
	pool := &WorkerPool{}

	app := NewAgentApp(client, dc, pool)

	if app.Client != client {
		t.Errorf("Expected client %v, got %v", client, app.Client)
	}
	if app.DataCollector != dc {
		t.Errorf("Expected data collector %v, got %v", dc, app.DataCollector)
	}
	if app.WorkerPool != pool {
		t.Errorf("Expected worker pool %v, got %v", pool, app.WorkerPool)
	}
}

func TestApp_handleResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pool := NewWorkerPool(1)
	app := &App{
		WorkerPool: pool,
	}

	go app.handleResults()

	pool.ResultCh <- Result{
		Err:      &url.Error{Err: os.ErrClosed},
		WorkerID: 1,
	}

	time.Sleep(100 * time.Millisecond)

	close(pool.ResultCh)
}

// TestApp_handleResultsNoError tests the handleResults function when no error occurs
func TestApp_handleResultsNoError(t *testing.T) {
	pool := NewWorkerPool(1)
	app := &App{
		WorkerPool: pool,
	}

	go app.handleResults()

	// Send a result without error
	pool.ResultCh <- Result{
		Value:    "test",
		Err:      nil,
		WorkerID: 1,
	}

	time.Sleep(100 * time.Millisecond)

	close(pool.ResultCh)
}
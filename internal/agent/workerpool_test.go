package agent

import (
	"github.com/rshafikov/alertme/internal/server/models"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNewWorkerPool(t *testing.T) {
	workers := 5
	wp := NewWorkerPool(workers)

	if wp.Workers != workers {
		t.Errorf("Expected %d workers, got %d", workers, wp.Workers)
	}

	if wp.JobsCh == nil {
		t.Error("Expected JobsCh to be initialized")
	}

	if wp.ResultCh == nil {
		t.Error("Expected ResultCh to be initialized")
	}

	if wp.DoneCh == nil {
		t.Error("Expected DoneCh to be initialized")
	}
}

func TestWorkerPool_Stop(t *testing.T) {
	wp := NewWorkerPool(1)
	
	// Stop should close the DoneCh channel
	wp.Stop()
	
	// Try to receive from the channel - it should not block
	select {
	case <-wp.DoneCh:
		// Channel is closed, which is expected
	case <-time.After(100 * time.Millisecond):
		t.Error("DoneCh channel was not closed after Stop()")
	}
}

func TestWorkerPool_RunWorkerStopSignal(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	testURL, _ := url.Parse(server.URL)
	client := NewClient(testURL)

	wp := NewWorkerPool(1)
	
	// Start worker in a goroutine
	go wp.RunWorker(1, client)
	
	// Give the worker time to start
	time.Sleep(100 * time.Millisecond)
	
	// Send stop signal
	wp.Stop()
	
	// Give the worker time to stop
	time.Sleep(100 * time.Millisecond)
}

func TestWorkerPool_RunWorkerJobProcessing(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	testURL, _ := url.Parse(server.URL)
	client := NewClient(testURL)

	wp := NewWorkerPool(1)
	
	// Start worker in a goroutine
	go wp.RunWorker(1, client)
	
	// Give the worker time to start
	time.Sleep(100 * time.Millisecond)
	
	// Send a job
	metricValue := 1.0
	job := []*models.Metric{
		{
			Name:  "test",
			Type:  models.GaugeType,
			Value: &metricValue,
		},
	}
	
	wp.JobsCh <- job
	
	// Wait for result
	select {
	case result := <-wp.ResultCh:
		if result.Err != nil {
			t.Errorf("Expected no error, got %v", result.Err)
		}
		if result.WorkerID != 1 {
			t.Errorf("Expected WorkerID 1, got %d", result.WorkerID)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for result")
	}
	
	// Clean up
	wp.Stop()
}

func TestWorkerPool_RunWorkerJobsChannelClosed(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	testURL, _ := url.Parse(server.URL)
	client := NewClient(testURL)

	wp := NewWorkerPool(1)
	
	// Start worker in a goroutine
	go wp.RunWorker(1, client)
	
	// Give the worker time to start
	time.Sleep(100 * time.Millisecond)
	
	// Close the jobs channel
	close(wp.JobsCh)
	
	// Give the worker time to process the close signal
	time.Sleep(100 * time.Millisecond)
}
package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogger(t *testing.T) {
	// Create a test handler that writes some data
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test data"))
	})

	// Wrap the handler with the Logger middleware
	loggedHandler := Logger(handler)

	// Create a request
	req := httptest.NewRequest("GET", "/", nil)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	loggedHandler.ServeHTTP(rr, req)

	// Check that the response is correct
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Body.String() != "test data" {
		t.Errorf("Expected response body to be 'test data', got '%s'", rr.Body.String())
	}
}
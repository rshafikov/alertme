package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGZipper(t *testing.T) {
	// Create a test handler that writes some data
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test data"))
	})

	// Wrap the handler with the GZipper middleware
	gzippedHandler := GZipper(handler)

	// Create a request with Accept-Encoding: gzip header
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	gzippedHandler.ServeHTTP(rr, req)

	// Check that the response is gzipped
	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding to be 'gzip', got '%s'", rr.Header().Get("Content-Encoding"))
	}

	// Check that the response body is gzipped
	gzReader, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	// Read the uncompressed data
	uncompressed, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("Failed to read uncompressed data: %v", err)
	}

	// Check that the uncompressed data is correct
	if string(uncompressed) != "test data" {
		t.Errorf("Expected uncompressed data to be 'test data', got '%s'", string(uncompressed))
	}
}

func TestGZipperNoGzip(t *testing.T) {
	// Create a test handler that writes some data
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test data"))
	})

	// Wrap the handler with the GZipper middleware
	gzippedHandler := GZipper(handler)

	// Create a request without Accept-Encoding: gzip header
	req := httptest.NewRequest("GET", "/", nil)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	gzippedHandler.ServeHTTP(rr, req)

	// Check that the response is not gzipped
	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Error("Expected Content-Encoding to not be 'gzip'")
	}

	// Check that the response body is not gzipped
	if rr.Body.String() != "test data" {
		t.Errorf("Expected response body to be 'test data', got '%s'", rr.Body.String())
	}
}

func TestGZipperCompressedRequest(t *testing.T) {
	// Create a test handler that reads the request body
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}
		w.Write(body)
	})

	// Wrap the handler with the GZipper middleware
	gzippedHandler := GZipper(handler)

	// Create a request with gzipped body
	var buf strings.Builder
	gzWriter := gzip.NewWriter(&buf)
	gzWriter.Write([]byte("test data"))
	gzWriter.Close()

	req := httptest.NewRequest("POST", "/", strings.NewReader(buf.String()))
	req.Header.Set("Content-Encoding", "gzip")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	gzippedHandler.ServeHTTP(rr, req)

	// Check that the response body is correct
	if rr.Body.String() != "test data" {
		t.Errorf("Expected response body to be 'test data', got '%s'", rr.Body.String())
	}
}
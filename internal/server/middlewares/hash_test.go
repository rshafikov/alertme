package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/rshafikov/alertme/internal/server/settings"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHasher(t *testing.T) {
	// Save original key
	originalKey := settings.CONF.Key
	defer func() {
		settings.CONF.Key = originalKey
	}()

	// Set a test key
	settings.CONF.Key = "testkey"

	// Create a test handler that writes some data
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test data"))
	})

	// Wrap the handler with the Hasher middleware
	hashedHandler := Hasher(handler)

	// Create a request
	req := httptest.NewRequest("GET", "/", nil)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	hashedHandler.ServeHTTP(rr, req)

	// Check that the response has a hash header
	hashHeader := rr.Header().Get("HashSHA256")
	if hashHeader == "" {
		t.Error("Expected HashSHA256 header to be set")
	}

	// Verify the hash
	expectedHash := hashData([]byte("test data"))
	expectedHashStr := hex.EncodeToString(expectedHash)
	if hashHeader != expectedHashStr {
		t.Errorf("Expected hash to be %s, got %s", expectedHashStr, hashHeader)
	}
}

func TestHasherWithValidHash(t *testing.T) {
	// Save original key
	originalKey := settings.CONF.Key
	defer func() {
		settings.CONF.Key = originalKey
	}()

	// Set a test key
	settings.CONF.Key = "testkey"

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test data"))
	})

	// Wrap the handler with the Hasher middleware
	hashedHandler := Hasher(handler)

	// Create a request with a valid hash
	body := []byte("test data")
	hash := hmac.New(sha256.New, []byte("testkey"))
	hash.Write(body)
	hashStr := hex.EncodeToString(hash.Sum(nil))

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("HashSHA256", hashStr)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	hashedHandler.ServeHTTP(rr, req)

	// Check that the response is successful
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestHasherWithInvalidHash(t *testing.T) {
	// Save original key
	originalKey := settings.CONF.Key
	defer func() {
		settings.CONF.Key = originalKey
	}()

	// Set a test key
	settings.CONF.Key = "testkey"

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test data"))
	})

	// Wrap the handler with the Hasher middleware
	hashedHandler := Hasher(handler)

	// Create a request with an invalid hash
	body := []byte("test data")
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("HashSHA256", "invalidhash")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	hashedHandler.ServeHTTP(rr, req)

	// Check that the response is a bad request
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHasherWithoutKey(t *testing.T) {
	// Save original key
	originalKey := settings.CONF.Key
	defer func() {
		settings.CONF.Key = originalKey
	}()

	// Set an empty key
	settings.CONF.Key = ""

	// Create a test handler that writes some data
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test data"))
	})

	// Wrap the handler with the Hasher middleware
	hashedHandler := Hasher(handler)

	// Create a request
	req := httptest.NewRequest("GET", "/", nil)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	hashedHandler.ServeHTTP(rr, req)

	// Check that the response does not have a hash header
	hashHeader := rr.Header().Get("HashSHA256")
	if hashHeader != "" {
		t.Errorf("Expected HashSHA256 header to be empty, got %s", hashHeader)
	}
}
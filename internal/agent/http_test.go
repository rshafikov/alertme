package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/rshafikov/alertme/internal/agent/config"
	"github.com/rshafikov/alertme/internal/server/models"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNewClient(t *testing.T) {
	testURL, _ := url.Parse("http://localhost:8080")
	client := NewClient(testURL)

	if client.URL != testURL {
		t.Errorf("Expected URL %v, got %v", testURL, client.URL)
	}
}

func TestClient_compressData(t *testing.T) {
	client := &Client{}
	data := []byte("test data")

	compressed, err := client.compressData(data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Decompress to verify
	reader, err := gzip.NewReader(bytes.NewReader(compressed.Bytes()))
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	var decompressed bytes.Buffer
	_, err = decompressed.ReadFrom(reader)
	if err != nil {
		t.Fatalf("Failed to decompress data: %v", err)
	}

	if !bytes.Equal(decompressed.Bytes(), data) {
		t.Errorf("Expected %s, got %s", string(data), decompressed.String())
	}
}

func TestClient_hashData(t *testing.T) {
	// Save original key
	originalKey := config.Key
	config.Key = "testkey"
	defer func() { config.Key = originalKey }()

	client := &Client{}
	data := []byte("test data")
	expectedHash := hmac.New(sha256.New, []byte(config.Key))
	expectedHash.Write(data)
	expected := hex.EncodeToString(expectedHash.Sum(nil)[:])

	result := client.hashData(data)
	if result != expected {
		t.Errorf("Expected hash %s, got %s", expected, result)
	}
}

func TestClient_hashDataEmptyKey(t *testing.T) {
	// Save original key
	originalKey := config.Key
	config.Key = ""
	defer func() { config.Key = originalKey }()

	client := &Client{}
	data := []byte("test data")
	result := client.hashData(data)
	
	// With empty key, should still produce a hash
	if result == "" {
		t.Error("Expected non-empty hash with empty key")
	}
}

func TestClient_sendMetricsContextCancelled(t *testing.T) {
	client := &Client{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	metrics := []*models.Metric{}
	err := client.sendMetrics(ctx, metrics)
	
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestClient_SendDataSuccess(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	testURL, _ := url.Parse(server.URL)
	client := NewClient(testURL)

	metricValue := 1.0
	metrics := []*models.Metric{
		{
			Name:  "test",
			Type:  models.GaugeType,
			Value: &metricValue,
		},
	}

	err := client.SendData(metrics)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
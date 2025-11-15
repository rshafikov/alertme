package logger

import (
	"testing"
)

func TestInitialize(t *testing.T) {
	// Test with a valid log level
	err := Initialize("info")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test with an invalid log level
	err = Initialize("invalid")
	if err == nil {
		t.Error("Expected an error for invalid log level, got nil")
	}
}
package config

import (
	"net"
	"os"
	"testing"
)

func TestInitAgentConfiguration(t *testing.T) {
	originalEnv := Env
	originalServerAddress := ServerAddress
	originalReportInterval := ReportInterval
	originalPollInterval := PollInterval
	originalLogLevel := LogLevel
	originalKey := Key
	originalRateLimit := RateLimit

	defer func() {
		Env = originalEnv
		ServerAddress = originalServerAddress
		ReportInterval = originalReportInterval
		PollInterval = originalPollInterval
		LogLevel = originalLogLevel
		Key = originalKey
		RateLimit = originalRateLimit
	}()

	os.Setenv("ADDRESS", "localhost:9090")
	os.Setenv("REPORT_INTERVAL", "20")
	os.Setenv("POLL_INTERVAL", "5")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("KEY", "testkey")
	os.Setenv("RATE_LIMIT", "10")
	defer os.Unsetenv("ADDRESS")
	defer os.Unsetenv("REPORT_INTERVAL")
	defer os.Unsetenv("POLL_INTERVAL")
	defer os.Unsetenv("LOG_LEVEL")
	defer os.Unsetenv("KEY")
	defer os.Unsetenv("RATE_LIMIT")

	ServerAddress = netAddress{Host: defaultHost, Port: defaultPort}
	ReportInterval = defaultReportInterval
	PollInterval = defaultPollInterval
	LogLevel = defaultLogLevel
	Key = ""
	RateLimit = defaultRateLimit

	if err := ParseEnv(); err != nil {
		t.Fatalf("Failed to parse environment variables: %v", err)
	}

	if Env.SrvAddr != "" {
		host, port, err := net.SplitHostPort(Env.SrvAddr)
		if err == nil {
			ServerAddress.Host = host
			ServerAddress.Port = port
		}
	}

	if Env.ReportIntrv > 0 {
		ReportInterval = Env.ReportIntrv
	}

	if Env.PollIntrv > 0 {
		PollInterval = Env.PollIntrv
	}

	if Env.LogLevel != "" {
		LogLevel = Env.LogLevel
	}

	if Env.Key != "" {
		Key = Env.Key
	}

	if Env.RateLimit > 0 {
		RateLimit = Env.RateLimit
	}

	if ServerAddress.Host != "localhost" {
		t.Errorf("Expected ServerAddress.Host to be 'localhost', got '%s'", ServerAddress.Host)
	}
	if ServerAddress.Port != "9090" {
		t.Errorf("Expected ServerAddress.Port to be '9090', got '%s'", ServerAddress.Port)
	}
	if ReportInterval != 20 {
		t.Errorf("Expected ReportInterval to be 20, got %d", ReportInterval)
	}
	if PollInterval != 5 {
		t.Errorf("Expected PollInterval to be 5, got %d", PollInterval)
	}
	if LogLevel != "debug" {
		t.Errorf("Expected LogLevel to be 'debug', got '%s'", LogLevel)
	}
	if Key != "testkey" {
		t.Errorf("Expected Key to be 'testkey', got '%s'", Key)
	}
	if RateLimit != 10 {
		t.Errorf("Expected RateLimit to be 10, got %d", RateLimit)
	}
}
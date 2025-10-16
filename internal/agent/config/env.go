package config

import (
	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
	"log"
)

type envConfig struct {
	LogLevel    string `env:"LOG_LEVEL"`
	SrvAddr     string `env:"ADDRESS"`
	Key         string `env:"KEY"`
	ReportIntrv int    `env:"REPORT_INTERVAL"`
	PollIntrv   int    `env:"POLL_INTERVAL"`
	RateLimit   int    `env:"RATE_LIMIT"`
}

// Env holds the configuration values loaded from environment variables.
var Env envConfig

// ParseEnv parses environment variables into the Env variable.
// Returns an error if parsing fails.
func ParseEnv() error {
	err := env.Parse(&Env)
	if err != nil {
		log.Println("Unable to parse ENV:", zap.Error(err))
		return err
	}
	return nil
}

package settings

import (
	"github.com/caarlos0/env/v6"
	"log"
)

type envServerConfig struct {
	ServerAddress   string `env:"ADDRESS"`
	StoreInteval    int    `env:"STORE_INTERVAL" envDefault:"-1"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	LogLevel        string `env:"LOG_LEVEL"`
	DatabaseURL     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
}

// ServerEnv holds the configuration values loaded from environment variables.
var ServerEnv envServerConfig

// ParseEnv parses environment variables into the ServerEnv variable.
// Returns an error if parsing fails.
func ParseEnv() error {
	err := env.Parse(&ServerEnv)
	if err != nil {
		log.Println("Unable to parse ENV:", err)
		return err
	}
	return nil
}

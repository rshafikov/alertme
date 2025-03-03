package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/rshafikov/alertme/internal/server/logger"
	"go.uber.org/zap"
)

type envServerConfig struct {
	ServerAddress   string `env:"ADDRESS"`
	StoreInteval    int    `env:"STORE_INTERVAL" envDefault:"-1"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	LogLevel        string `env:"LOG_LEVEL"`
}

var ServerEnv envServerConfig

func ParseEnv() error {
	err := env.Parse(&ServerEnv)
	if err != nil {
		logger.Log.Error("Unable to parse ENV:", zap.Error(err))
		return err
	}
	return nil
}

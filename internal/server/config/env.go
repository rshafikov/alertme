package config

import (
	"github.com/caarlos0/env/v6"
)

type envServerConfig struct {
	ServerAddress   string `env:"ADDRESS"`
	StoreInteval    int    `env:"STORE_INTERVAL" envDefault:"-1"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

var ServerEnv envServerConfig

func ParseEnv() error {
	err := env.Parse(&ServerEnv)
	if err != nil {
		Log.Errorln("Unable to parse ENV:", err.Error())
		return err
	}
	return nil
}

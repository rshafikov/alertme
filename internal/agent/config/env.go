package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/rshafikov/alertme/internal/server/config"
)

type envConfig struct {
	SrvAddr     string `env:"ADDRESS"`
	ReportIntrv int    `env:"REPORT_INTERVAL"`
	PollIntrv   int    `env:"POLL_INTERVAL"`
}

var Env envConfig

func ParseEnv() error {
	err := env.Parse(&Env)
	if err != nil {
		config.Log.Errorln("Unable to parse ENV:", err.Error())
		return err
	}
	return nil
}

package config

import (
	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
	"log"
)

type envConfig struct {
	SrvAddr     string `env:"ADDRESS"`
	ReportIntrv int    `env:"REPORT_INTERVAL"`
	PollIntrv   int    `env:"POLL_INTERVAL"`
	LogLevel    string `env:"LOG_LEVEL"`
}

var Env envConfig

func ParseEnv() error {
	err := env.Parse(&Env)
	if err != nil {
		log.Println("Unable to parse ENV:", zap.Error(err))
		return err
	}
	return nil
}

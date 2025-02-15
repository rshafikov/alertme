package config

import (
	"fmt"
	"github.com/caarlos0/env/v6"
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
		fmt.Println("Unable to parse ENV:", err)
		return err
	}
	return nil
}

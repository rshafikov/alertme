package agent

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"net"
	"strconv"
)

type EnvConfig struct {
	SrvAddr     string `env:"ADDRESS"`
	ReportIntrv int    `env:"REPORT_INTERVAL"`
	PollIntrv   int    `env:"POLL_INTERVAL"`
}

type netAddress struct {
	Host string
	Port string
}

func (na *netAddress) String() string {
	return fmt.Sprintf("%s:%s", na.Host, na.Port)
}

func (na *netAddress) Set(value string) error {
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return err
	}
	_, err = strconv.Atoi(port)
	if err != nil {
		return err
	}
	na.Host = host
	na.Port = port
	return nil
}

var ServerAddress = netAddress{Host: "localhost", Port: "8080"}
var ReportInterval int = 10
var PollInterval int = 2

var Env EnvConfig

func ParseEnv() error {
	err := env.Parse(&Env)
	if err != nil {
		fmt.Println("Unable to parse ENV:", err)
		return err
	}
	return nil
}

func InitAgentFlags() {
	flag.Var(&ServerAddress, "a", "server address")
	flag.IntVar(&ReportInterval, "r", ReportInterval, "report interval")
	flag.IntVar(&PollInterval, "p", PollInterval, "poll interval")
	flag.Parse()
}

func InitAgentConfiguration() {
	InitAgentFlags()
	if err := ParseEnv(); err == nil {
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
	}
}

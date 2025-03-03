package config

import (
	"flag"
	"fmt"
	"github.com/rshafikov/alertme/internal/server/config"
	"log"
	"net"
	"strconv"
)

const (
	defaultHost           = "localhost"
	defaultPort           = "8080"
	defaultReportInterval = 10
	defaultPollInterval   = 2
)

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

var ServerAddress = netAddress{Host: defaultHost, Port: defaultPort}
var ReportInterval int
var PollInterval int

func InitAgentFlags() {
	flag.Var(&ServerAddress, "a", "server address")
	flag.IntVar(&ReportInterval, "r", defaultReportInterval, "report interval")
	flag.IntVar(&PollInterval, "p", defaultPollInterval, "poll interval")
	flag.Parse()
	if ReportInterval <= 0 {
		log.Fatal("report interval cannot be negative or null")
	}
	if PollInterval <= 0 {
		log.Fatal("poll interval cannot be negative or null")
	}
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
	config.Log.Infof("\n"+
		"\033[1;36m╭────────────────────────────────────────\033[0m\n"+
		"\033[1;36m│ \033[1;34m🚀 Agent Initialized Successfully \033[1;36m\033[0m\n"+
		"\033[1;36m├────────────────────────────────────────\033[0m\n"+
		"\033[1;37m│ \033[1;33m📡 Server Address:   \033[0;37m%-47s \033[1;36m\033[0m\n"+
		"\033[1;37m│ \033[1;33m⏱  Report Interval:  \033[0;37m%-47d \033[1;36m\033[0m\n"+
		"\033[1;37m│ \033[1;33m⏱  Poll interval:    \033[0;37m%-47d \033[1;36m\033[0m\n"+
		"\033[1;36m╰────────────────────────────────────────\033[0m",
		ServerAddress.String(), ReportInterval, PollInterval,
	)
}

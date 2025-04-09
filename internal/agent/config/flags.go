package config

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
)

const (
	defaultHost           = "localhost"
	defaultPort           = "8080"
	defaultReportInterval = 10
	defaultPollInterval   = 2
	defaultLogLevel       = "info"
	defaultRateLimit      = 2
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
var LogLevel string
var Key string
var RateLimit int

func InitAgentFlags() {
	flag.Var(&ServerAddress, "a", "server address")
	flag.IntVar(&ReportInterval, "r", defaultReportInterval, "report interval")
	flag.IntVar(&PollInterval, "p", defaultPollInterval, "poll interval")
	flag.StringVar(&LogLevel, "v", defaultLogLevel, "log level")
	flag.StringVar(&Key, "k", "", "key to sign sending data")
	flag.IntVar(&RateLimit, "l", defaultRateLimit, "rate limit")
	flag.Parse()

	if ReportInterval <= 0 {
		log.Fatal("report interval cannot be negative or null")
	}

	if PollInterval <= 0 {
		log.Fatal("poll interval cannot be negative or null")
	}
}

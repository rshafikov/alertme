package agent

import (
	"flag"
	"fmt"
	"net"
	"strconv"
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

var ServerAddress = netAddress{Host: "localhost", Port: "8080"}
var ReportInterval int = 10
var PollInterval int = 2

func InitAgentFlags() {
	flag.Var(&ServerAddress, "a", "server address")
	flag.IntVar(&ReportInterval, "r", ReportInterval, "report interval")
	flag.IntVar(&PollInterval, "p", PollInterval, "poll interval")
	flag.Parse()
}

package config

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	baseHost     = "localhost"
	baseHostPort = "8080"
)

type netAddress struct {
	Host string
	Port string
}

func (na *netAddress) String() string {
	return fmt.Sprintf("%s:%s", na.Host, na.Port)
}

func (na *netAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}
	_, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	na.Host = hp[0]
	na.Port = hp[1]
	return nil
}

var Address = netAddress{Host: baseHost, Port: baseHostPort}

func InitServerFlags() {
	_ = flag.Value(&Address)
	flag.Var(&Address, "a", "server address")
	flag.Parse()

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		host, port, err := net.SplitHostPort(envAddress)
		if err != nil {
			fmt.Printf("invalid ADDRESS environment variable: %s\n", envAddress)
			os.Exit(1)
		}
		Address.Host = host
		Address.Port = port
	}
}

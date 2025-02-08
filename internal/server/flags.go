package server

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
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

var Address netAddress = netAddress{Host: "localhost", Port: "8080"}

func InitServerFlags() {
	_ = flag.Value(&Address)
	flag.Var(&Address, "a", "server address")
	flag.Parse()
}

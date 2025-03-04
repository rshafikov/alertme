package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	baseHost            = "localhost"
	baseHostPort        = "8080"
	baseStoreInterval   = 300
	baseFileStoragePath = "metrics.txt"
	baseRestore         = false
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
var StoreInterval int
var FileStoragePath string
var Restore bool

func InitServerFlags() {
	_ = flag.Value(&Address)
	flag.Var(&Address, "a", "server address")
	flag.IntVar(&StoreInterval, "i", baseStoreInterval, "interval to store metrics, in seconds")
	flag.StringVar(&FileStoragePath, "f", baseFileStoragePath, "storage path - file to store metrics")
	flag.BoolVar(&Restore, "r", baseRestore, "restore metrics from file, specified in the storage path")
	flag.Parse()

	if StoreInterval < 0 {
		log.Fatal("store interval cannot be negative")
	}
}

func InitServerConfiguration() {
	InitServerFlags()

	if err := ParseEnv(); err == nil {
		if ServerEnv.ServerAddress != "" {
			host, port, err := net.SplitHostPort(ServerEnv.ServerAddress)
			if err != nil {
				log.Fatalln("invalid ADDRESS environment variable:", ServerEnv.ServerAddress)
			}
			Address.Host = host
			Address.Port = port
		}
		if ServerEnv.StoreInteval >= 0 {
			StoreInterval = ServerEnv.StoreInteval
		}
		if ServerEnv.FileStoragePath != "" {
			FileStoragePath = ServerEnv.FileStoragePath
		}
		if ServerEnv.Restore {
			Restore = ServerEnv.Restore
		}
	}

	Log.Infof("\n"+
		"\033[1;36m╭────────────────────────────────────────\033[0m\n"+
		"\033[1;36m│ \033[1;34m🚀 Server Initialized Successfully \033[1;36m\033[0m\n"+
		"\033[1;36m├────────────────────────────────────────\033[0m\n"+
		"\033[1;37m│ \033[1;33m📡 Address:         \033[0;37m%-47s \033[1;36m\033[0m\n"+
		"\033[1;37m│ \033[1;33m⏱  Store Interval:  \033[0;37m%-47d \033[1;36m\033[0m\n"+
		"\033[1;37m│ \033[1;33m💾 Storage Path:    \033[0;37m%-47s \033[1;36m\033[0m\n"+
		"\033[1;37m│ \033[1;33m🔄 Restore State:   \033[0;37m%-47t \033[1;36m\033[0m\n"+
		"\033[1;36m╰────────────────────────────────────────\033[0m",
		Address.String(), StoreInterval, FileStoragePath, Restore,
	)
}

package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
)

const (
	defaultHost            = "localhost"
	defaultHostPort        = "8080"
	defaultStoreInterval   = 300
	defaultFileStoragePath = "metrics.txt"
	defaultRestore         = false
	defaultLogLevel        = "info"
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
		return errors.New("supported format: host:port")
	}
	_, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	na.Host = hp[0]
	na.Port = hp[1]
	return nil
}

type dbSettings struct {
	Driver   string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// postgres://postgres:password@localhost:5432/postgres
func (dbu *dbSettings) String() string {
	url := fmt.Sprintf("%s://%s:%s@%s:%s/%s", dbu.Driver, dbu.User, dbu.Password, dbu.Host, dbu.Port, dbu.Name)
	if url == "://:@:/" {
		url = ""
	}
	return url
}

func (dbu *dbSettings) Set(s string) error {
	parsed := strings.Split(s, "://")
	if len(parsed) != 2 || parsed[0] == "" {
		return errors.New("supported format: postgres://user:password@host:port/dbname")
	}

	dbu.Driver = parsed[0]
	credentialsHostDB := parsed[1]
	atIndex := strings.LastIndex(credentialsHostDB, "@")
	if atIndex == -1 {
		return errors.New("invalid url: missing '@'")
	}

	credentials := credentialsHostDB[:atIndex]
	hostDB := credentialsHostDB[atIndex+1:]

	creds := strings.SplitN(credentials, ":", 2)
	if len(creds) != 2 {
		return errors.New("invalid url: missing or invalid credentials")
	}
	dbu.User = creds[0]
	dbu.Password = creds[1]

	slashIndex := strings.LastIndex(hostDB, "/")
	if slashIndex == -1 {
		return errors.New("invalid url: missing '/' before dbname")
	}
	hostPort := hostDB[:slashIndex]
	dbu.Name = hostDB[slashIndex+1:]

	colonIndex := strings.LastIndex(hostPort, ":")
	if colonIndex == -1 {
		return errors.New("invalid url: missing ':' in host:port")
	}
	dbu.Host = hostPort[:colonIndex]
	dbu.Port = hostPort[colonIndex+1:]

	return nil
}

var Address = netAddress{Host: defaultHost, Port: defaultHostPort}
var DatabaseSettings = dbSettings{}
var StoreInterval int
var FileStoragePath string
var Restore bool
var LogLevel string
var DatabaseURL string

func InitServerFlags() {
	_ = flag.Value(&Address)
	flag.Var(&Address, "a", "server address")

	_ = flag.Value(&DatabaseSettings)
	flag.Var(&DatabaseSettings, "d", "database url")

	flag.IntVar(&StoreInterval, "i", defaultStoreInterval, "interval to store metrics, in seconds")
	flag.StringVar(&FileStoragePath, "f", defaultFileStoragePath, "storage path - file to store metrics")
	flag.BoolVar(&Restore, "r", defaultRestore, "restore metrics from file, specified in the storage path")
	flag.StringVar(&LogLevel, "l", defaultLogLevel, "log level")
	flag.Parse()

	if StoreInterval < 0 {
		log.Fatal("store interval cannot be negative")
	}

	DatabaseURL = DatabaseSettings.String()
}

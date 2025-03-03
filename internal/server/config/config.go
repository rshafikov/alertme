package config

import (
	"fmt"
	"log"
	"net"
)

func InitServerConfiguration() {
	InitServerFlags()

	if err := ParseEnv(); err == nil {
		if ServerEnv.ServerAddress != "" {
			host, port, err := net.SplitHostPort(ServerEnv.ServerAddress)
			if err != nil {
				log.Fatal("invalid ADDRESS environment variable: ", ServerEnv.ServerAddress)
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
		if ServerEnv.LogLevel != "" {
			LogLevel = ServerEnv.LogLevel
		}
		if ServerEnv.DatabaseURL != "" {
			err := DatabaseSettings.Set(ServerEnv.DatabaseURL)
			if err != nil {
				log.Fatal("Unable to parse DATABASE_DSN environment variable: ", ServerEnv.DatabaseURL)
			}
			DatabaseURL = DatabaseSettings.String()
		}
	}
	initMessage := "\033[1;36m╭────────────────────────────────────────\033[0m\n" +
		"\033[1;36m│ \033[1;34m🚀 Server Initialized Successfully \033[0m\n" +
		"\033[1;36m├────────────────────────────────────────\033[0m\n" +
		"\033[1;36m│ \033[1;33m📡 Address:         \033[0;37m%-39s\033[0m\n" +
		"\033[1;36m│ \033[1;33m⏱️ Store Interval:  \033[0;37m%-39d\033[0m\n" +
		"\033[1;36m│ \033[1;33m💾 Storage Path:    \033[0;37m%-39s\033[0m\n" +
		"\033[1;36m│ \033[1;33m🔄 Restore State:   \033[0;37m%-39t\033[0m\n" +
		"\033[1;36m│ \033[1;33m🗂️Database DSN:    \033[0;37m%-39s\033[0m\n" +
		"\033[1;36m╰────────────────────────────────────────\033[0m\n"

	dbURLMessage := "-------"
	if DatabaseURL != "" {
		dbURLMessage = DatabaseURL
	}
	fmt.Printf(initMessage, Address.String(), StoreInterval, FileStoragePath, Restore, dbURLMessage)
}

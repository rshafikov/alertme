package settings

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
			CONF.ServerAddress.Host = host
			CONF.ServerAddress.Port = port
		}
		if ServerEnv.StoreInteval >= 0 {
			CONF.StoreInterval = ServerEnv.StoreInteval
		}
		if ServerEnv.FileStoragePath != "" {
			CONF.FileStoragePath = ServerEnv.FileStoragePath
		}
		if ServerEnv.Restore {
			CONF.Restore = ServerEnv.Restore
		}
		if ServerEnv.LogLevel != "" {
			CONF.LogLevel = ServerEnv.LogLevel
		}
		if ServerEnv.DatabaseURL != "" {
			err := CONF.DatabaseSettings.Set(ServerEnv.DatabaseURL)
			if err != nil {
				log.Fatal("Unable to parse DATABASE_DSN environment variable: ", ServerEnv.DatabaseURL)
			}
			CONF.DatabaseURL = CONF.DatabaseSettings.String()
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
	if CONF.DatabaseURL != "" {
		dbURLMessage = CONF.DatabaseURL
	}
	fmt.Printf(
		initMessage,
		CONF.ServerAddress.String(),
		CONF.StoreInterval,
		CONF.FileStoragePath,
		CONF.Restore,
		dbURLMessage,
	)
}

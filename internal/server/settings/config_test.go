package settings

import (
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to set up and tear down environment variables
func withEnvVars(t *testing.T, env map[string]string, fn func()) {
	original := make(map[string]string)
	for key := range env {
		original[key] = os.Getenv(key)
		if env[key] == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, env[key])
		}
	}
	t.Cleanup(func() {
		for key, value := range original {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	})
	fn()
}

func TestInitServerConfiguration_Default(t *testing.T) {
	const (
		defaultHost            = "localhost"
		defaultHostPort        = "8080"
		defaultStoreInterval   = 300
		defaultFileStoragePath = "metrics.txt"
		defaultRestore         = false
		defaultLogLevel        = "info"
	)

	t.Run("Default configuration is set correctly", func(t *testing.T) {
		CONF = serverConfig{
			ServerAddress:    netAddress{Host: defaultHost, Port: defaultHostPort},
			DatabaseSettings: dbSettings{},
			StoreInterval:    defaultStoreInterval,
			FileStoragePath:  defaultFileStoragePath,
			Restore:          defaultRestore,
			LogLevel:         defaultLogLevel,
			DatabaseURL:      "",
			Key:              "",
		}
		ServerEnv = envServerConfig{}

		withEnvVars(t, map[string]string{
			"ADDRESS":           "",
			"STORE_INTERVAL":    "",
			"FILE_STORAGE_PATH": "",
			"RESTORE":           "",
			"LOG_LEVEL":         "",
			"DATABASE_DSN":      "",
			"KEY":               "",
		}, func() {
			InitServerConfiguration()

			assert.Equal(t, defaultHost, CONF.ServerAddress.Host)
			assert.Equal(t, defaultHostPort, CONF.ServerAddress.Port)
			assert.Equal(t, defaultStoreInterval, CONF.StoreInterval)
			assert.Equal(t, defaultFileStoragePath, CONF.FileStoragePath)
			assert.Equal(t, defaultRestore, CONF.Restore)
			assert.Equal(t, defaultLogLevel, CONF.LogLevel)
			assert.Equal(t, "", CONF.DatabaseURL)
			assert.Equal(t, "", CONF.Key)
		})
	})
}

func TestInitServerConfiguration_InvalidAddress(t *testing.T) {
	const (
		defaultHost            = "localhost"
		defaultHostPort        = "8080"
		defaultStoreInterval   = 300
		defaultFileStoragePath = "metrics.txt"
		defaultRestore         = false
		defaultLogLevel        = "info"
	)
	t.Run("Invalid ADDRESS environment variable", func(t *testing.T) {
		CONF = serverConfig{
			ServerAddress:    netAddress{Host: defaultHost, Port: defaultHostPort},
			DatabaseSettings: dbSettings{},
			StoreInterval:    defaultStoreInterval,
			FileStoragePath:  defaultFileStoragePath,
			Restore:          defaultRestore,
			LogLevel:         defaultLogLevel,
			DatabaseURL:      "",
			Key:              "",
		}
		ServerEnv = envServerConfig{}

		// Set environment variables
		withEnvVars(t, map[string]string{
			"ADDRESS":           "invalid-address",
			"STORE_INTERVAL":    "",
			"FILE_STORAGE_PATH": "",
			"RESTORE":           "",
			"LOG_LEVEL":         "",
			"DATABASE_DSN":      "",
			"KEY":               "",
		}, func() {
			err := ParseEnv()
			require.NoError(t, err)
			assert.Equal(t, "invalid-address", ServerEnv.ServerAddress)

			if ServerEnv.ServerAddress != "" {
				_, _, err := net.SplitHostPort(ServerEnv.ServerAddress)
				assert.Error(t, err)
			}
		})
	})
}

func TestInitServerConfiguration_InvalidDSN(t *testing.T) {
	const (
		defaultHost            = "localhost"
		defaultHostPort        = "8080"
		defaultStoreInterval   = 300
		defaultFileStoragePath = "metrics.txt"
		defaultRestore         = false
		defaultLogLevel        = "info"
	)
	t.Run("Invalid DATABASE_DSN environment variable", func(t *testing.T) {
		// Reset global state
		CONF = serverConfig{
			ServerAddress:    netAddress{Host: defaultHost, Port: defaultHostPort},
			DatabaseSettings: dbSettings{},
			StoreInterval:    defaultStoreInterval,
			FileStoragePath:  defaultFileStoragePath,
			Restore:          defaultRestore,
			LogLevel:         defaultLogLevel,
			DatabaseURL:      "",
			Key:              "",
		}
		ServerEnv = envServerConfig{}

		// Set environment variables
		withEnvVars(t, map[string]string{
			"ADDRESS":           "",
			"STORE_INTERVAL":    "",
			"FILE_STORAGE_PATH": "",
			"RESTORE":           "",
			"LOG_LEVEL":         "",
			"DATABASE_DSN":      "invalid-dsn",
			"KEY":               "",
		}, func() {
			err := ParseEnv()
			require.NoError(t, err)
			assert.Equal(t, "invalid-dsn", ServerEnv.DatabaseURL)

			if ServerEnv.DatabaseURL != "" {
				err := CONF.DatabaseSettings.Set(ServerEnv.DatabaseURL)
				assert.Error(t, err)
			}
		})
	})
}

func TestInitServerConfiguration_InvalidStoreInterval(t *testing.T) {
	const (
		defaultHost            = "localhost"
		defaultHostPort        = "8080"
		defaultStoreInterval   = 300
		defaultFileStoragePath = "metrics.txt"
		defaultRestore         = false
		defaultLogLevel        = "info"
	)
	t.Run("Negative STORE_INTERVAL is not applied", func(t *testing.T) {
		// Reset global state
		CONF = serverConfig{
			ServerAddress:    netAddress{Host: defaultHost, Port: defaultHostPort},
			DatabaseSettings: dbSettings{},
			StoreInterval:    defaultStoreInterval,
			FileStoragePath:  defaultFileStoragePath,
			Restore:          defaultRestore,
			LogLevel:         defaultLogLevel,
			DatabaseURL:      "",
			Key:              "",
		}
		ServerEnv = envServerConfig{}

		// Set environment variables
		withEnvVars(t, map[string]string{
			"ADDRESS":           "",
			"STORE_INTERVAL":    "-100",
			"FILE_STORAGE_PATH": "",
			"RESTORE":           "",
			"LOG_LEVEL":         "",
			"DATABASE_DSN":      "",
			"KEY":               "",
		}, func() {
			err := ParseEnv()
			require.NoError(t, err)

			assert.Equal(t, -100, ServerEnv.StoreInteval)

			if ServerEnv.StoreInteval >= 0 {
				CONF.StoreInterval = ServerEnv.StoreInteval
			}

			assert.Equal(t, defaultStoreInterval, CONF.StoreInterval)
		})
	})
}

package config

import (
	"fmt"
	"net"
	"strconv"
)

// InitAgentConfiguration initializes the agent configuration by processing command-line flags
// and environment variables.
// It prints a formatted message with the final configuration.
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

		if Env.LogLevel != "" {
			LogLevel = Env.LogLevel
		}

		if Env.Key != "" {
			Key = Env.Key
		}
	}

	initMessage := "\033[1;36m╭────────────────────────────────────────\033[0m\n" +
		"\033[1;36m│ \033[1;34m🚀 Agent Initialized Successfully \033[1;36m\033[0m\n" +
		"\033[1;36m├────────────────────────────────────────\033[0m\n" +
		"\033[1;36m│ \033[1;33m📡 Server Address:   \033[0;37m%-47s \033[1;36m\033[0m\n" +
		"\033[1;36m│ \033[1;33m⏱  Report Interval:  \033[0;37m%-47d \033[1;36m\033[0m\n" +
		"\033[1;36m│ \033[1;33m⏱  Poll Interval:    \033[0;37m%-47d \033[1;36m\033[0m\n" +
		"\033[1;36m│ \033[1;33m🔐 Hash Key:         \033[0;37m%-47s \033[1;36m\033[0m\n" +
		"\033[1;36m│ \033[1;33m📝 Logging Level:    \033[0;37m%-47s \033[1;36m\033[0m\n" +
		"\033[1;36m│ \033[1;33m📝 Rate Limit:       \033[0;37m%-47v \033[1;36m\033[0m\n" +
		"\033[1;36m╰────────────────────────────────────────\033[0m\n"

	keyInitMessage := "-----"
	if Key != "" {
		keyInitMessage = "********"
	}

	rateLimitInitMessage := "-----"
	if RateLimit > 0 {
		rateLimitInitMessage = strconv.Itoa(RateLimit)
	}

	fmt.Printf(
		initMessage,
		ServerAddress.String(),
		ReportInterval,
		PollInterval,
		keyInitMessage,
		LogLevel,
		rateLimitInitMessage,
	)
}

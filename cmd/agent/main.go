package main

import (
	"github.com/rshafikov/alertme/internal/agent"
	"github.com/rshafikov/alertme/internal/agent/config"
	"github.com/rshafikov/alertme/internal/agent/metrics"
	"github.com/rshafikov/alertme/internal/server/logger"
	"log"
	"net/url"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	config.InitAgentConfiguration()

	if err := logger.Initialize(config.LogLevel); err != nil {
		log.Fatal(err)
	}
	printBuildInfo()

	baseURL, err := url.Parse("http://" + config.ServerAddress.String())
	if err != nil {
		log.Fatal(err)
	}

	dc := metrics.NewEmptyDataCollector()
	client := agent.NewClient(baseURL)
	wp := agent.NewWorkerPool(config.RateLimit)

	app := agent.NewAgentApp(client, dc, wp)
	app.Start()
}

func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	log.Printf("Build version: %s\n", buildVersion)
	log.Printf("Build date: %s\n", buildDate)
	log.Printf("Build commit: %s\n", buildCommit)
}

package main

import (
	"github.com/rshafikov/alertme/internal/agent"
	"github.com/rshafikov/alertme/internal/agent/config"
	"github.com/rshafikov/alertme/internal/agent/metrics"
	"github.com/rshafikov/alertme/internal/server/logger"
	"log"
	"net/url"
)

func main() {
	config.InitAgentConfiguration()

	if err := logger.Initialize(config.LogLevel); err != nil {
		log.Fatal(err)
	}

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

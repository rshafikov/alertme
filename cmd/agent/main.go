package main

import (
	"fmt"
	"github.com/rshafikov/alertme/internal/agent"
	"github.com/rshafikov/alertme/internal/agent/config"
	"github.com/rshafikov/alertme/internal/agent/metrics"
	"github.com/rshafikov/alertme/internal/server/logger"
	"log"
	"net/url"
	"time"
)

func main() {
	config.InitAgentConfiguration()
	if err := logger.Initialize(config.LogLevel); err != nil {
		log.Fatal(err)
	}

	dc := metrics.NewEmptyDataCollector()

	baseURL, err := url.Parse("http://" + config.ServerAddress.String())
	if err != nil {
		log.Fatal(err)
	}

	client := agent.NewClient(baseURL)
	pollTicker := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)

	go func() {
		for range pollTicker.C {
			CollectMetrics(dc)
		}
	}()

	go func() {
		for range reportTicker.C {
			SendMetrics(client, dc)
		}
	}()

	select {}
}

func CollectMetrics(dc *metrics.DataCollector) {
	dc.UpdateMetrics()
	fmt.Println("Metrics were updated")
}

func SendMetrics(client *agent.Client, dataCollector *metrics.DataCollector) {
	client.SendStoredData(dataCollector)
	fmt.Println("Metrics were sent")
}

package main

import (
	"fmt"
	"github.com/rshafikov/alertme/internal/agent"
	"github.com/rshafikov/alertme/internal/agent/config"
	"github.com/rshafikov/alertme/internal/agent/metrics"
	"time"
)

func main() {
	config.InitAgentConfiguration()
	dc := metrics.NewEmptyDataCollector()
	client := agent.NewClient("http://" + config.ServerAddress.String())
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
	fmt.Println(dc)
}

func SendMetrics(client *agent.Client, dataCollector *metrics.DataCollector) {
	client.SendStoredData(dataCollector)
	fmt.Println("Metrics were sent")
}

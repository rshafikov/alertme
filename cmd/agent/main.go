package main

import (
	"fmt"
	"github.com/rshafikov/alertme/internal/agent"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

func main() {
	CollectAndSendMetrics(reportInterval, pollInterval)
}

func CollectMetrics(dc *agent.DataCollector, interval time.Duration) {
	agent.UpdateDataCollector(dc)
	fmt.Println(dc)
	time.Sleep(interval)
}
func CollectAndSendMetrics(reportInterval, pollInterval time.Duration) {
	client := agent.NewClient("http://localhost:8080")
	dc := agent.NewEmptyDataCollector()
	timer := time.Now().Add(reportInterval)
	for {
		CollectMetrics(dc, pollInterval)
		if time.Now().After(timer) {
			client.SendStoredData(dc)
			fmt.Println("Metrics were sent")
			timer = time.Now().Add(reportInterval)
		}
	}
}

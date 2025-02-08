package main

import (
	"fmt"
	"github.com/rshafikov/alertme/internal/agent"
	"time"
)

func main() {
	agent.InitAgentFlags()
	CollectAndSendMetrics(agent.ReportInterval, agent.PollInterval)
}

func CollectMetrics(dc *agent.DataCollector, interval time.Duration) {
	agent.UpdateDataCollector(dc)
	fmt.Println(dc)
	time.Sleep(interval)
}

func CollectAndSendMetrics(reportTime, pollTime int) {
	client := agent.NewClient("http://" + agent.ServerAddress.String())
	dc := agent.NewEmptyDataCollector()
	timer := time.Now().Add(time.Duration(reportTime) * time.Second)
	for {
		CollectMetrics(dc, time.Duration(pollTime)*time.Second)
		if time.Now().After(timer) {
			client.SendStoredData(dc)
			fmt.Println("Metrics were sent")
			timer = time.Now().Add(time.Duration(reportTime) * time.Second)
		}
	}
}

package main

import (
	"github.com/rshafikov/alertme/internal/agent"
	"github.com/rshafikov/alertme/internal/agent/config"
	"github.com/rshafikov/alertme/internal/agent/metrics"
	"github.com/rshafikov/alertme/internal/server/logger"
	"go.uber.org/zap"
	"log"
	"net/url"
	"time"
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

	app := NewAgentApp(client, dc, wp)
	app.Start()
}

func CollectMetrics(dc *metrics.DataCollector) {
	dc.UpdateMetrics()
	logger.Log.Info("metrics were updated")
}

type AgentApp struct {
	Client        *agent.Client
	DataCollector *metrics.DataCollector
	WorkerPool    *agent.WorkerPool
}

func NewAgentApp(client *agent.Client, dc *metrics.DataCollector, pool *agent.WorkerPool) *AgentApp {
	return &AgentApp{
		Client:        client,
		DataCollector: dc,
		WorkerPool:    pool,
	}
}

func (app *AgentApp) Start() {
	pollTicker := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)

	go func() {
		for range pollTicker.C {
			logger.Log.Debug("collect metrics")
			CollectMetrics(app.DataCollector)
		}
	}()

	go func() {
		for range reportTicker.C {
			logger.Log.Debug("sending metrics to JobsCh")
			metricsToSend := append(app.DataCollector.Metrics, app.DataCollector.PollCount)
			app.WorkerPool.JobsCh <- metricsToSend
		}
	}()

	go func() {
		for range app.WorkerPool.ResultCh {
			logger.Log.Debug("handle ResultCh")
			r := <-app.WorkerPool.ResultCh
			if r.Err != nil {
				logger.Log.Error("worker failed to send metrics", zap.Int("worker_id", r.WorkerID), zap.Error(r.Err))
			}
		}
	}()

	for i := 1; i <= app.WorkerPool.Workers; i++ {
		go app.WorkerPool.RunWorker(i, app.Client)
	}

	select {}
}

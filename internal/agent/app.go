package agent

import (
	"github.com/rshafikov/alertme/internal/agent/config"
	"github.com/rshafikov/alertme/internal/agent/metrics"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/models"
	"go.uber.org/zap"
	"time"
)

type App struct {
	Client *Client
	DC     *metrics.DataCollector
	WP     *WorkerPool
}

func NewAgentApp(client *Client, dc *metrics.DataCollector, pool *WorkerPool) *App {
	return &App{
		Client: client,
		DC:     dc,
		WP:     pool,
	}
}

func (app *App) Start() {
	pollTicker := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
	shutdown := make(chan struct{})

	defer func() {
		pollTicker.Stop()
		reportTicker.Stop()
		close(shutdown)
	}()

	go app.collectMetrics(pollTicker)
	go app.reportMetrics(reportTicker)
	go app.handleResults()

	for i := 1; i <= app.WP.Workers; i++ {
		go app.WP.RunWorker(i, app.Client)
	}

	<-shutdown
}

func (app *App) collectMetrics(ticker *time.Ticker) {
	for range ticker.C {
		logger.Log.Debug("collecting metrics")
		go app.DC.UpdateRuntimeMetrics()
		go app.DC.UpdatePSUtilMetrics()
	}
}

func (app *App) sendMetrics(metrics []*models.Metric) {
	app.WP.JobsCh <- metrics
}

func (app *App) reportMetrics(ticker *time.Ticker) {
	for range ticker.C {
		logger.Log.Debug("sending metrics")
		app.sendMetrics(app.getRuntimeMetrics())
		app.sendMetrics(app.getPSUtilMetrics())
	}
}

func (app *App) handleResults() {
	for r := range app.WP.ResultCh {
		if r.Err != nil {
			logger.Log.Error("worker failed",
				zap.Int("worker_id", r.WorkerID),
				zap.Error(r.Err),
			)
		}
	}
}

func (app *App) getRuntimeMetrics() []*models.Metric {
	totalLen := len(app.DC.Metrics) + 1
	m := make([]*models.Metric, 0, totalLen)
	m = append(m, app.DC.Metrics...)
	m = append(m, app.DC.PollCount)
	return m
}

func (app *App) getPSUtilMetrics() []*models.Metric {
	totalLen := len(app.DC.CPUutilization) + 2
	m := make([]*models.Metric, 0, totalLen)
	m = append(m, app.DC.CPUutilization...)
	m = append(m, app.DC.TotalMemory, app.DC.FreeMemory)
	return m
}

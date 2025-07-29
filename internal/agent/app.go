package agent

import (
	"github.com/rshafikov/alertme/internal/agent/config"
	"github.com/rshafikov/alertme/internal/agent/metrics"
	"github.com/rshafikov/alertme/internal/server/logger"
	"go.uber.org/zap"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const profilingServerAddress = ":8888"

// App represents the agent application that collects and sends metrics.
type App struct {
	Client        *Client
	DataCollector *metrics.DataCollector
	WorkerPool    *WorkerPool
}

// NewAgentApp creates a new agent application with the provided client, data collector, and worker pool.
func NewAgentApp(client *Client, dc *metrics.DataCollector, pool *WorkerPool) *App {
	return &App{
		Client:        client,
		DataCollector: dc,
		WorkerPool:    pool,
	}
}

// Start begins the agent's operation, collecting and sending metrics at regular intervals.
// It sets up profiling if enabled, initializes tickers for polling and reporting,
// and handles graceful shutdown on interrupt signals.
func (app *App) Start() {
	if config.Profiling {
		logger.Log.Info("starting pprof server")
		go func() {
			err := http.ListenAndServe(profilingServerAddress, nil)
			if err != nil {
				logger.Log.Error(err.Error())
			}
		}()
	}

	pollTicker := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	sendTicker := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
	shutdown := make(chan os.Signal, 1)

	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	defer func() {
		pollTicker.Stop()
		sendTicker.Stop()
	}()

	go app.DataCollector.CollectMetrics(pollTicker)
	go app.DataCollector.SendMetrics(sendTicker, app.WorkerPool.JobsCh)
	go app.handleResults()

	for i := 1; i <= app.WorkerPool.Workers; i++ {
		go app.WorkerPool.RunWorker(i, app.Client)
	}

	<-shutdown
	logger.Log.Info("received shutdown signal, stopping workers...")
	app.WorkerPool.Stop()
	time.Sleep(300 * time.Millisecond)
}

func (app *App) handleResults() {
	for r := range app.WorkerPool.ResultCh {
		if r.Err != nil {
			logger.Log.Error("worker failed",
				zap.Int("worker_id", r.WorkerID),
				zap.Error(r.Err),
			)
		}
	}
}

package agent

import (
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/models"
	"go.uber.org/zap"
)

// Result represents the result of a worker's job.
type Result struct {
	Value    any
	Err      error
	WorkerID int
}

// WorkerPool represents a pool of workers that process jobs concurrently.
type WorkerPool struct {
	Workers  int
	JobsCh   chan []*models.Metric
	ResultCh chan Result
	DoneCh   chan struct{}
}

// NewWorkerPool creates a new worker pool with the specified number of workers.
func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		Workers:  workers,
		JobsCh:   make(chan []*models.Metric),
		ResultCh: make(chan Result),
		DoneCh:   make(chan struct{}),
	}
}

// Stop signals all workers to stop processing jobs.
func (wp *WorkerPool) Stop() {
	close(wp.DoneCh)
}

// RunWorker starts a worker with the specified ID that processes jobs from the JobsCh.
// The worker sends metrics to the server using the provided client and reports results to ResultCh.
func (wp *WorkerPool) RunWorker(id int, client *Client) {
	logger.Log.Debug("worker starting", zap.Int("worker_id", id))

	for {
		select {

		case <-wp.DoneCh:
			logger.Log.Debug("worker recieved stop signal", zap.Int("worker_id", id))
			return

		case v, ok := <-wp.JobsCh:
			if !ok {
				logger.Log.Debug("jobs channel closed, closing worker", zap.Int("worker_id", id))
				return
			}
			logger.Log.Debug("worker received job", zap.Int("worker_id", id))

			err := client.SendData(v)

			if err != nil {
				logger.Log.Debug("an error occurred while processing job", zap.Error(err))
			}

			wp.ResultCh <- Result{Value: nil, Err: err, WorkerID: id}
		}
	}
}

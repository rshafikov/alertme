package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/models"
	"go.uber.org/zap"
	"os"
	"time"
)

// FileSaver provides methods to persist and restore metrics storage to and from a file.
// Storage is the metrics storage interface to manage metric operations.
// FileName specifies the file path used for saving and loading metrics.
type FileSaver struct {
	Storage  BaseMetricStorage
	FileName string
}

// NewFileSaver initializes a new FileSaver with the given storage and file path.
func NewFileSaver(storage BaseMetricStorage, filePath string) FileSaver {
	return FileSaver{
		Storage:  storage,
		FileName: filePath,
	}
}

// LoadMetrics reads metrics from a file and returns them as a slice of Metric instances.
func (l *FileSaver) LoadMetrics() ([]*models.Metric, error) {
	file, err := os.Open(l.FileName)
	if err != nil {
		logger.Log.Error(errmsg.UnableToOpenFile, zap.Error(err))
		return nil, err
	}
	defer file.Close()

	var fileMetrics []*models.Metric

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		var metric models.Metric
		if jsonErr := json.Unmarshal([]byte(line), &metric); jsonErr != nil {
			continue
		}
		fileMetrics = append(fileMetrics, &metric)

	}

	return fileMetrics, nil
}

// SaveMetrics writes a slice of Metric instances to a file in JSON format.
// It creates or overwrites the specified file and ignores encoding errors for individual metrics.
// Returns an error if the file cannot be opened or written to.
func (l *FileSaver) SaveMetrics(metrics []*models.Metric) error {
	file, err := os.OpenFile(l.FileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Log.Error(errmsg.UnableToOpenFile, zap.Error(err))
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, metric := range metrics {
		if encodeErr := encoder.Encode(metric); encodeErr != nil {
			continue
		}
	}

	return nil
}

// LoadStorage restores metrics from a file and populates the in-memory storage.
// Returns an error if loading or storage operations fail.
func (l *FileSaver) LoadStorage(ctx context.Context) error {
	oldMetrics, loadErr := l.LoadMetrics()
	if loadErr != nil {
		return loadErr
	}
	for _, oldMetric := range oldMetrics {
		err := l.Storage.Add(ctx, oldMetric)
		if err != nil {
			logger.Log.Error("unable to restore metric", zap.Error(err))
			return err
		}
	}

	logger.Log.Info("Storage was restored")
	return nil
}

// SaveStorage saves in-memory metrics to a file. Returns an error if saving fails.
func (l *FileSaver) SaveStorage(ctx context.Context) error {
	logger.Log.Debug("trying to save metrics to", zap.String("filename", l.FileName))
	err := l.SaveMetrics(l.Storage.List(ctx))

	if err != nil {
		return errors.New(errmsg.UnableToAddMetric)
	}
	logger.Log.Debug("metrics successfully saved to", zap.String("filename", l.FileName))
	return nil
}

// SaveStorageWithInterval periodically saves storage content at the specified interval in seconds.
func (l *FileSaver) SaveStorageWithInterval(ctx context.Context, interval int) error {
	if interval < 0 {
		return errors.New("interval must be a positive int value")
	}

	if l.Storage == nil {
		return errors.New("storage is nil")
	}

	storeTicker := time.NewTicker(time.Duration(interval) * time.Second)

	go func() {
		select {
		case <-storeTicker.C:
			_ = l.SaveStorage(ctx)
		case <-ctx.Done():
			storeTicker.Stop()
			return
		}
	}()
	return nil
}

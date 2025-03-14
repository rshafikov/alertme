package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/models"
	"go.uber.org/zap"
	"os"
	"time"
)

func NewFileSaver(storage BaseMetricStorage, filePath string) FileSaver {
	return FileSaver{
		Storage:  storage,
		FileName: filePath,
	}
}

type FileSaver struct {
	Storage  BaseMetricStorage
	FileName string
}

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

func (l *FileSaver) LoadStorage() error {
	oldMetrics, loadErr := l.LoadMetrics()
	if loadErr != nil {
		return loadErr
	}
	for _, oldMetric := range oldMetrics {
		err := l.Storage.Add(oldMetric)
		if err != nil {
			logger.Log.Error(errmsg.UnableToRestoreMetric, zap.Error(err))
			return err
		}
		logger.Log.Info("Storage was restored")
	}
	return nil
}

func (l *FileSaver) SaveStorage() error {
	logger.Log.Debug("trying to save metrics to", zap.String("filename", l.FileName))
	err := l.SaveMetrics(l.Storage.List())
	if err != nil {
		return errors.New(errmsg.UnableToSaveMetricInStorage)
	}
	logger.Log.Debug("metrics successfully saved to", zap.String("filename", l.FileName))
	return nil
}

func (l *FileSaver) SaveStorageWithInterval(interval int) error {
	if interval < 0 {
		return errors.New(errmsg.IntervalMustBePositive)
	}

	if l.Storage == nil {
		return errors.New(errmsg.StorageIsNil)
	}

	storeTicker := time.NewTicker(time.Duration(interval) * time.Second)

	go func() {
		for range storeTicker.C {
			_ = l.SaveStorage()
		}
	}()
	return nil
}

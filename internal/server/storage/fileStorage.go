package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/rshafikov/alertme/internal/server/config"
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/models"
	"os"
	"time"
)

func NewFileSaver(filePath string) FileLoader {
	return FileLoader{FileName: filePath}
}

type FileLoader struct {
	FileName string
}

func (l *FileLoader) LoadMetrics() ([]*models.Metric, error) {
	file, err := os.Open(l.FileName)
	if err != nil {
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

func (l *FileLoader) SaveMetrics(metrics []*models.Metric) error {
	file, err := os.OpenFile(l.FileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
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

func (l *FileLoader) LoadStorage(storage BaseMetricStorage) error {
	oldMetrics, loadErr := l.LoadMetrics()
	if loadErr != nil {
		return loadErr
	}
	for _, oldMetric := range oldMetrics {
		err := storage.Add(oldMetric)
		if err != nil {
			config.Log.Errorln("Unable to add old metric to storage", err.Error())
			return err
		}
	}
	return nil
}

func (l *FileLoader) SaveStorage(storage BaseMetricStorage) error {
	config.Log.Debugf("trying to save metrics to %s...", l.FileName)
	err := l.SaveMetrics(storage.List())
	if err != nil {
		return errors.New(errmsg.UnableToSaveMetricInStorage)
	}
	config.Log.Debugf("metrics successfully saved to %s", l.FileName)
	return nil
}

func (l *FileLoader) SaveStorageWithInterval(interval int, storage BaseMetricStorage) error {
	if interval < 0 {
		return errors.New(errmsg.IntervalMustBePositive)
	}

	if storage == nil {
		return errors.New(errmsg.StorageIsNil)
	}

	storeTicker := time.NewTicker(time.Duration(interval) * time.Second)

	go func() {
		for range storeTicker.C {
			_ = l.SaveStorage(storage)
		}
	}()
	return nil
}

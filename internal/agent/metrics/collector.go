package metrics

import (
	"context"
	"fmt"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
	"math/rand"
	"runtime"
	"time"
)

const SendMetricsTimeout = time.Second * 1

type MetricSource int

const (
	RuntimeMetrics MetricSource = iota
	PSUtilMetrics
)

type DataCollector struct {
	Metrics        []*models.Metric
	PollCount      *models.Metric
	TotalMemory    *models.Metric
	FreeMemory     *models.Metric
	CPUUtilization []*models.Metric
}

func NewEmptyDataCollector() *DataCollector {
	initalCount := int64(0)
	return &DataCollector{
		PollCount: &models.Metric{Name: "PollCount", Delta: &initalCount, Type: models.CounterType},
	}
}

func (d *DataCollector) CollectMetrics(ticker *time.Ticker) {
	for range ticker.C {
		logger.Log.Debug("collecting metrics")
		d.UpdateRuntimeMetrics()
		d.UpdatePSUtilMetrics()
	}
}

func (d *DataCollector) String() string {
	metrics := "========================================\n"
	for _, metric := range d.Metrics {
		metrics += fmt.Sprintf("%v\n", metric)
	}
	metrics += fmt.Sprintf("========================================\n%v", d.PollCount)
	return metrics
}

func (d *DataCollector) UpdateRuntimeMetrics() {
	if d.PollCount.Delta == nil {
		d.PollCount.Delta = new(int64)
	}
	*d.PollCount.Delta++

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	d.Metrics = []*models.Metric{
		{Name: "Alloc", Type: models.GaugeType, Value: float64Ptr(float64(memStats.Alloc))},
		{Name: "BuckHashSys", Type: models.GaugeType, Value: float64Ptr(float64(memStats.BuckHashSys))},
		{Name: "Frees", Type: models.GaugeType, Value: float64Ptr(float64(memStats.Frees))},
		{Name: "GCCPUFraction", Type: models.GaugeType, Value: float64Ptr(memStats.GCCPUFraction)},
		{Name: "GCSys", Type: models.GaugeType, Value: float64Ptr(float64(memStats.GCSys))},
		{Name: "HeapAlloc", Type: models.GaugeType, Value: float64Ptr(float64(memStats.HeapAlloc))},
		{Name: "HeapIdle", Type: models.GaugeType, Value: float64Ptr(float64(memStats.HeapIdle))},
		{Name: "HeapInuse", Type: models.GaugeType, Value: float64Ptr(float64(memStats.HeapInuse))},
		{Name: "HeapObjects", Type: models.GaugeType, Value: float64Ptr(float64(memStats.HeapObjects))},
		{Name: "HeapReleased", Type: models.GaugeType, Value: float64Ptr(float64(memStats.HeapReleased))},
		{Name: "HeapSys", Type: models.GaugeType, Value: float64Ptr(float64(memStats.HeapSys))},
		{Name: "LastGC", Type: models.GaugeType, Value: float64Ptr(float64(memStats.LastGC))},
		{Name: "Lookups", Type: models.GaugeType, Value: float64Ptr(float64(memStats.Lookups))},
		{Name: "MCacheInuse", Type: models.GaugeType, Value: float64Ptr(float64(memStats.MCacheInuse))},
		{Name: "MCacheSys", Type: models.GaugeType, Value: float64Ptr(float64(memStats.MCacheSys))},
		{Name: "MSpanInuse", Type: models.GaugeType, Value: float64Ptr(float64(memStats.MSpanInuse))},
		{Name: "MSpanSys", Type: models.GaugeType, Value: float64Ptr(float64(memStats.MSpanSys))},
		{Name: "Mallocs", Type: models.GaugeType, Value: float64Ptr(float64(memStats.Mallocs))},
		{Name: "NextGC", Type: models.GaugeType, Value: float64Ptr(float64(memStats.NextGC))},
		{Name: "NumForcedGC", Type: models.GaugeType, Value: float64Ptr(float64(memStats.NumForcedGC))},
		{Name: "NumGC", Type: models.GaugeType, Value: float64Ptr(float64(memStats.NumGC))},
		{Name: "OtherSys", Type: models.GaugeType, Value: float64Ptr(float64(memStats.OtherSys))},
		{Name: "PauseTotalNs", Type: models.GaugeType, Value: float64Ptr(float64(memStats.PauseTotalNs))},
		{Name: "StackInuse", Type: models.GaugeType, Value: float64Ptr(float64(memStats.StackInuse))},
		{Name: "StackSys", Type: models.GaugeType, Value: float64Ptr(float64(memStats.StackSys))},
		{Name: "Sys", Type: models.GaugeType, Value: float64Ptr(float64(memStats.Sys))},
		{Name: "TotalAlloc", Type: models.GaugeType, Value: float64Ptr(float64(memStats.TotalAlloc))},
		{Name: "RandomValue", Type: models.GaugeType, Value: float64Ptr(rand.Float64())},
	}
}

func (d *DataCollector) UpdatePSUtilMetrics() {
	memoryData, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Error("Failed to get virtual memory info", zap.Error(err))
	}

	d.TotalMemory = &models.Metric{
		Name:  "TotalMemory",
		Value: float64Ptr(float64(memoryData.Total)),
		Delta: nil,
		Type:  models.GaugeType,
	}

	d.FreeMemory = &models.Metric{
		Name:  "FreeMemory",
		Value: float64Ptr(float64(memoryData.Free)),
		Delta: nil,
		Type:  models.GaugeType,
	}
	cpuData, err := cpu.Percent(time.Second, true)
	if err != nil {
		logger.Log.Error("Failed to get cpu usage", zap.Error(err))
		return
	}

	var cpuMetrics []*models.Metric
	for c, v := range cpuData {
		cpuMetrics = append(
			cpuMetrics,
			&models.Metric{
				Name:  fmt.Sprintf("CPUutilization%v", c),
				Value: float64Ptr(v),
				Delta: nil,
				Type:  models.GaugeType,
			},
		)
	}

	d.CPUUtilization = cpuMetrics
}

func (d *DataCollector) PassMetrics(t MetricSource, ch chan []*models.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), SendMetricsTimeout)
	defer cancel()

	var metrics []*models.Metric
	switch t {
	case RuntimeMetrics:
		metrics = append(metrics, d.Metrics...)
		metrics = append(metrics, d.PollCount)

	case PSUtilMetrics:
		metrics = append(metrics, d.CPUUtilization...)
		metrics = append(metrics, d.TotalMemory, d.FreeMemory)
	}

	select {
	case ch <- metrics:
	case <-ctx.Done():
		logger.Log.Warn("metrics weren't sent, reached timeout", zap.Duration("timeout", SendMetricsTimeout))
	}

}

func (d *DataCollector) SendMetrics(ticker *time.Ticker, jobs chan []*models.Metric) {
	for range ticker.C {
		logger.Log.Debug("collecting metrics")
		go d.PassMetrics(RuntimeMetrics, jobs)
		go d.PassMetrics(PSUtilMetrics, jobs)
	}
}

func float64Ptr(f float64) *float64 {
	return &f
}

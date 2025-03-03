package metrics

import (
	"fmt"
	"github.com/rshafikov/alertme/internal/server/models"
	"math/rand"
	"runtime"
)

type DataCollector struct {
	Metrics   []*models.Metric
	PollCount *models.Metric
}

func NewEmptyDataCollector() *DataCollector {
	initalCount := int64(0)
	return &DataCollector{
		PollCount: &models.Metric{Name: "PollCount", Delta: &initalCount, Type: models.CounterType},
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

func (d *DataCollector) UpdateMetrics() {
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

func float64Ptr(f float64) *float64 {
	return &f
}

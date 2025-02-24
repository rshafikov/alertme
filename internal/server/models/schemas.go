package models

import "strconv"

type MetricJSONReq struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (metric *MetricJSONReq) ConvertToBaseMetric(anotherMetric *Metric) *Metric {
	metric.ID = anotherMetric.Name
	metric.MType = string(anotherMetric.Type)
	switch anotherMetric.Type {
	case CounterType:
		delta, _ := strconv.ParseInt(anotherMetric.Value, 10, 64)
		metric.Delta = &delta
	case GaugeType:
		value, _ := strconv.ParseFloat(anotherMetric.Value, 64)
		metric.Value = &value
	}
	return anotherMetric
}

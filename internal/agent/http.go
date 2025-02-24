package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rshafikov/alertme/internal/agent/metrics"
	"github.com/rshafikov/alertme/internal/server/models"
	"net/http"
	"net/url"
)

type Client struct {
	URL *url.URL
}

func NewClient(serverURL *url.URL) *Client {
	return &Client{URL: serverURL}
}

func (c *Client) SendStoredData(data *metrics.DataCollector) {
	for _, m := range data.Metrics {

		c.sendMetric(&models.MetricJSONReq{
			ID:    m.Name,
			MType: string(m.Type),
			Value: &m.Value,
		})
	}
	c.sendMetric(&models.MetricJSONReq{
		ID:    data.PollCount.Name,
		MType: string(data.PollCount.Type),
		Delta: &data.PollCount.Value,
	})
}

func (c *Client) sendMetric(metric *models.MetricJSONReq) {
	jsonBody, err := json.Marshal(metric)
	if err != nil {
		fmt.Println("failed to serialize metric:", err)
	}
	resp, reqErr := http.Post(c.URL.String()+"/update/", "application/json", bytes.NewBuffer(jsonBody))
	if reqErr != nil {
		fmt.Println("failed to send metric:", reqErr)
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("failed to send metric:", resp.Status)
	}

	defer resp.Body.Close()
}

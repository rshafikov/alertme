package agent

import (
	"fmt"
	"github.com/rshafikov/alertme/internal/agent/metrics"
	"github.com/rshafikov/alertme/internal/server/models"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

type Client struct {
	URL string
}

func NewClient(url string) *Client {
	return &Client{URL: url}
}

func (c *Client) SendStoredData(data *metrics.DataCollector) {
	for _, m := range data.Metrics {
		c.sendMetric(m.Type, m.Name, strconv.FormatFloat(m.Value, 'f', -1, 64))
	}
	c.sendMetric(data.PollCount.Type, data.PollCount.Name, strconv.FormatInt(data.PollCount.Value, 10))
}

func (c *Client) sendMetric(metricType models.MetricType, metricName, metricValue string) {
	baseURL, err := url.Parse(c.URL)
	if err != nil {
		fmt.Println("invalid base URL:", err)
		return
	}
	baseURL.Path = path.Join(baseURL.Path, "update", string(metricType), metricName, metricValue)

	resp, err := http.Post(baseURL.String(), "text/plain", nil)
	if err != nil {
		fmt.Println("failed to send metric:", err)
		return
	}
	defer resp.Body.Close()
}

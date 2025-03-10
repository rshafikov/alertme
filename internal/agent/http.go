package agent

import (
	"bytes"
	"compress/gzip"
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
		c.sendMetric(m)
	}
	c.sendMetric(data.PollCount)
}

func (c *Client) sendMetric(metric *models.Metric) {
	jsonBody, err := json.Marshal(metric)
	if err != nil {
		fmt.Println("failed to serialize metric:", err)
	}

	gzipData, err := c.compressMetric(jsonBody)
	if err != nil {
		fmt.Println("failed to compress metric:", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, c.URL.String()+"/update/", gzipData)
	if err != nil {
		fmt.Println("failed to create request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("failed to send request:", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("failed to send metric:", resp.Status)
		return
	}

	defer resp.Body.Close()
}

func (c *Client) compressMetric(data []byte) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	_, err := zb.Write(data)
	if err != nil {
		return nil, err
	}
	err = zb.Close()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

package agent

import (
	"fmt"
	"github.com/rshafikov/alertme/internal/server/models"
	"net/http"
	"strconv"
	"strings"
)

type Client struct {
	URL string
}

func NewClient(url string) *Client {
	return &Client{URL: url}
}

func (c *Client) SendStoredData(data *DataCollector) {
	for _, m := range data.Metrics {
		c.sendMetric(m.Type, m.Name, strconv.FormatFloat(m.Value, 'f', -1, 64))
	}
	c.sendMetric(data.PollCount.Type, data.PollCount.Name, strconv.FormatInt(data.PollCount.Value, 10))
}

func (c *Client) sendMetric(t models.MetricType, n, v string) {
	url := c.URL + fmt.Sprintf("/update/%v/%v/%v", t, n, v)
	resp, err := http.Post(url, "text/plain", strings.NewReader(""))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
}

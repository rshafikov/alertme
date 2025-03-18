package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"github.com/rshafikov/alertme/internal/agent/metrics"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/models"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"time"
)

var ErrUnableToSendMetrics = errors.New("unable to send metrics")

type Client struct {
	URL *url.URL
}

func NewClient(serverURL *url.URL) *Client {
	return &Client{URL: serverURL}
}

func (c *Client) SendStoredData(data *metrics.DataCollector) error {
	metricsToSend := append(data.Metrics, data.PollCount)
	err := c.sendMetrics(context.Background(), metricsToSend)

	if err != nil && errors.Is(err, ErrUnableToSendMetrics) {
		expTimeouts := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
		logger.Log.Warn("unable to send metrics, retrying", zap.Error(err))

		for i, timeout := range expTimeouts {
			time.Sleep(timeout)

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			err = c.sendMetrics(ctx, metricsToSend)
			if err == nil {
				logger.Log.Info("metrics sent successfully on retry", zap.Int("attempt", i+1))
				return nil
			}

			if ctx.Err() != nil {
				logger.Log.Warn("retry attempt failed due to timeout", zap.Int("attempt", i+1))
			} else {
				logger.Log.Error("retry attempt to send metrics failed", zap.Int("attempt", i+1))
			}
		}
		return err
	}
	return nil
}

func (c *Client) sendMetrics(ctx context.Context, metric []*models.Metric) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	jsonBody, err := json.Marshal(metric)
	if err != nil {
		logger.Log.Error("failed to serialize metrics:", zap.Error(err))
		return err
	}

	gzipData, err := c.compressMetric(jsonBody)
	if err != nil {
		logger.Log.Error("failed to compress metric:", zap.Error(err))
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL.String()+"/updates/", gzipData)
	if err != nil {
		logger.Log.Error("failed to create request:", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log.Error("failed to send request:", zap.Error(err))
		return ErrUnableToSendMetrics
	}
	if resp.StatusCode == http.StatusInternalServerError {
		logger.Log.Error("internal server error", zap.Error(ErrUnableToSendMetrics))
		return ErrUnableToSendMetrics
	}

	defer resp.Body.Close()
	return nil
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

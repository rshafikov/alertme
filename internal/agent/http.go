package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/rshafikov/alertme/internal/agent/config"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/rshafikov/alertme/internal/server/retry"
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

func (c *Client) SendData(metrics []*models.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := retry.OnErr(ctx, []error{ErrUnableToSendMetrics}, []time.Duration{
		1 * time.Second, 3 * time.Second, 5 * time.Second},
		func(args ...any) error {
			return c.sendMetrics(ctx, metrics)
		},
	)

	if err != nil {
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

	gzipData, err := c.compressData(jsonBody)
	if err != nil {
		logger.Log.Error("failed to compress metric:", zap.Error(err))
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL.String()+"/updates/", gzipData)
	if err != nil {
		logger.Log.Error("failed to create request:", zap.Error(err))
		return err
	}

	if config.Key != "" {
		hash := c.hashData(jsonBody)
		req.Header.Set("HashSHA256", hash)
		logger.Log.Info("hash:", zap.String("hash", hash))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log.Error("failed to send request:", zap.Error(err))
		return ErrUnableToSendMetrics
	}

	if resp.StatusCode != http.StatusOK {
		logger.Log.Error(
			"unable to send metrics",
			zap.Int("response_code", resp.StatusCode),
			zap.Error(ErrUnableToSendMetrics),
		)
		return ErrUnableToSendMetrics
	}

	defer resp.Body.Close()
	return nil
}

func (c *Client) compressData(data []byte) (*bytes.Buffer, error) {
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

func (c *Client) hashData(data []byte) string {
	h := hmac.New(sha256.New, []byte(config.Key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)[:])
}

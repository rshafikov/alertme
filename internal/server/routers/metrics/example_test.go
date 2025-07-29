package metrics

import (
	"context"
	"fmt"
	"github.com/rshafikov/alertme/internal/server/models"
	"github.com/rshafikov/alertme/internal/server/storage"
	"io"
	"net/http"
	"net/http/httptest"
)

func ExampleRouter_CreateMetricFromURL() {
	s := storage.NewMemStorage()
	r := NewMetricsRouter(s)
	ts := httptest.NewServer(r.Routes())
	defer ts.Close()

	post, err := http.Post(ts.URL+"/update/gauge/importantMetric/1337", "", nil)
	if err != nil {
		panic(err)
	}
	defer post.Body.Close()

	fmt.Printf("Status: %d\n", post.StatusCode)
	// Output: Status: 200
}

func ExampleRouter_GetMetricFromURL() {
	s := storage.NewMemStorage()
	r := NewMetricsRouter(s)
	ts := httptest.NewServer(r.Routes())
	defer ts.Close()

	metricVal := 13.37

	s.Add(context.Background(), &models.Metric{
		Name:  "testMetric",
		Value: &metricVal,
		Delta: nil,
		Type:  models.GaugeType,
	})

	get, err := http.Get(ts.URL + "/value/gauge/testMetric")
	if err != nil {
		panic(err)
	}
	defer get.Body.Close()

	body, _ := io.ReadAll(get.Body)

	fmt.Printf("metric %s", body)
	// Output: metric 13.37
}

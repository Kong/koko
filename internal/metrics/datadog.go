package metrics

import (
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-go/v5/statsd"
	"go.uber.org/zap"
)

type datadogClient struct {
	client statsd.ClientInterface
}

const (
	defaultRate = 1
)

func newDatadogClient(agentAddr string) (*datadogClient, error) {
	client, err := statsd.New(agentAddr, statsd.WithNamespace(metricNamespace))
	if err != nil {
		return nil, err
	}
	return &datadogClient{client: client}, nil
}

func (c *datadogClient) Gauge(name string, value float64, tags ...Tag) error {
	return c.client.Gauge(name, value, convertTags(tags...), defaultRate)
}

func (c *datadogClient) Count(name string, value int64, tags ...Tag) error {
	return c.client.Count(name, value, convertTags(tags...), defaultRate)
}

func (c *datadogClient) Histogram(name string, value float64, tags ...Tag) error {
	return c.client.Histogram(name, value, convertTags(tags...), defaultRate)
}

func (c *datadogClient) CreateHandler(log *zap.Logger) http.Handler {
	return nil
}

func (c *datadogClient) Close() {
	c.client.Close()
}

func convertTags(tags ...Tag) []string {
	ddtags := make([]string, 0, len(tags))
	for _, tag := range tags {
		ddtags = append(ddtags, fmt.Sprintf("%s:%s", tag.Name, tag.Value))
	}
	return ddtags
}

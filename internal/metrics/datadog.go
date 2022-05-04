package metrics

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-go/v5/statsd"
	"go.uber.org/zap"
)

type datadogClient struct {
	log    *zap.Logger
	client statsd.ClientInterface
}

const (
	defaultRate = 1
)

func newDatadogClient(logger *zap.Logger, agentAddr string) (*datadogClient, error) {
	client, err := statsd.New(agentAddr)
	if err != nil {
		return nil, err
	}
	return &datadogClient{log: logger, client: client}, nil
}

func (c *datadogClient) Gauge(name string, value float64, tags ...Tag) {
	if err := c.client.Gauge(name, value, convertTags(tags...), defaultRate); err != nil {
		c.log.With(zap.Error(err)).Error("failed to update gauge", zap.String("name", name))
	}
}

func (c *datadogClient) Count(name string, value int64, tags ...Tag) {
	if err := c.client.Count(name, value, convertTags(tags...), defaultRate); err != nil {
		c.log.With(zap.Error(err)).Error("failed to update count", zap.String("name", name))
	}
}

func (c *datadogClient) Histogram(name string, value float64, tags ...Tag) {
	if err := c.client.Histogram(name, value, convertTags(tags...), defaultRate); err != nil {
		c.log.With(zap.Error(err)).Error("failed to update histogram", zap.String("name", name))
	}
}

func (c *datadogClient) CreateHandler(log *zap.Logger) (http.Handler, error) {
	return nil, errors.New("datadog metrics client has no http.Handler")
}

func (c *datadogClient) Close() error {
	return c.client.Close()
}

func convertTags(tags ...Tag) []string {
	ddtags := make([]string, 0, len(tags))
	for _, tag := range tags {
		ddtags = append(ddtags, fmt.Sprintf("%s:%s", tag.Key, tag.Value))
	}
	return ddtags
}

package metrics

import (
	"errors"
	"net/http"

	"go.uber.org/zap"
)

type noopClient struct{}

func (c noopClient) Gauge(string, float64, ...Tag) {}

func (c noopClient) GaugeAdd(string, float64, ...Tag) {}

func (c noopClient) Count(string, int64, ...Tag) {}

func (c noopClient) Histogram(string, float64, ...Tag) {}

func (c noopClient) CreateHandler(log *zap.Logger) (http.Handler, error) {
	return nil, errors.New("noop metrics client has no http.Handler")
}

func (c noopClient) Close() error {
	return nil
}

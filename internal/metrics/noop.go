package metrics

import (
	"net/http"

	"go.uber.org/zap"
)

type noopClient struct{}

func (c noopClient) Gauge(name string, value float64, tags ...Tag) error {
	return nil
}

func (c noopClient) Count(name string, value int64, tags ...Tag) error {
	return nil
}

func (c noopClient) Histogram(name string, value float64, tags ...Tag) error {
	return nil
}

func (c noopClient) CreateHandler(log *zap.Logger) http.Handler {
	return nil
}

func (c noopClient) Close() {}

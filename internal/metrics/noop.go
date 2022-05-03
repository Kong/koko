package metrics

import (
	"errors"
	"net/http"

	"go.uber.org/zap"
)

type noopClient struct{}

func (c noopClient) Gauge(name string, value float64, tags ...Tag) {}

func (c noopClient) Count(name string, value int64, tags ...Tag) {}

func (c noopClient) Histogram(name string, value float64, tags ...Tag) {}

func (c noopClient) CreateHandler(log *zap.Logger) (http.Handler, error) {
	return nil, errors.New("noop metrics client has no http.Handler")
}

func (c noopClient) Close() {}

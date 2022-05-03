package metrics

import (
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"
)

var activeClient metricsClient = noopClient{}

type Tag struct {
	Key  string
	Value string
}

type metricsClient interface {
	Gauge(name string, value float64, tags ...Tag) error
	Count(name string, value int64, tags ...Tag) error
	Histogram(name string, value float64, tags ...Tag) error
	CreateHandler(log *zap.Logger) http.Handler
	Close()
}

type ClientType int

const (
	NoOp ClientType = iota
	Datadog
	Prometheus
)

const (
	metricNamespace = "koko"
)

var validClientTypes = map[string]ClientType{
	"noop":       NoOp,
	"datadog":    Datadog,
	"prometheus": Prometheus,
}

func ParseClientType(clientType string) (ClientType, error) {
	if clientType == "" {
		return NoOp, nil
	}

	if c, ok := validClientTypes[clientType]; ok {
		return c, nil
	}
	return NoOp, fmt.Errorf("invalid metrics_client '%s'", clientType)
}

func (c ClientType) String() string {
	for k, client := range validClientTypes {
		if client == c {
			return k
		}
	}
	panic("invalid client")
}

func InitMetricsClient(logger *zap.Logger, clientType string) error {
	ct, err := ParseClientType(clientType)
	if err != nil {
		return err
	}

	switch ct {
	case Datadog:
		agent := os.Getenv("DD_AGENT_HOST")
		if agent == "" {
			panic("Datadog client environment variable 'DD_AGENT_HOST' must be set")
		}

		var err error
		activeClient, err = newDatadogClient(agent)
		if err != nil {
			return err
		}
	case Prometheus:
		activeClient = newPrometheusClient()
	case NoOp:
		fallthrough
	default:
		logger.Info("metrics client config not set")
	}
	return nil
}

func Gauge(name string, value float64, tags ...Tag) error {
	return activeClient.Gauge(name, value, tags...)
}

func Count(name string, value int64, tags ...Tag) error {
	return activeClient.Count(name, value, tags...)
}

func Histogram(name string, value float64, tags ...Tag) error {
	return activeClient.Histogram(name, value, tags...)
}

func CreateHandler(log *zap.Logger) http.Handler {
	return activeClient.CreateHandler(log)
}

func Close() {
	activeClient.Close()
}

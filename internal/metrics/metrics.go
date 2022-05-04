package metrics

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"
)

var activeClient metricsClient = noopClient{}

type Tag struct {
	Key   string
	Value string
}

type metricsClient interface {
	// Gauge measures the value of a metric at a particular time.
	Gauge(name string, value float64, tags ...Tag)

	// Count tracks how many times something happened.
	Count(name string, value int64, tags ...Tag)

	// Histogram tracks the statistical distribution of a set of values.
	Histogram(name string, value float64, tags ...Tag)

	// CreateHandler create an http.Handler if supported by the client.
	// Otherwise an error will be returned.
	CreateHandler(log *zap.Logger) (http.Handler, error)

	// Close the underlying client connection if supported.
	Close() error
}

type ClientType int

const (
	NoOp ClientType = iota
	Datadog
	Prometheus
)

const (
	metricPrefix = "kong"
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
	return NoOp, fmt.Errorf("invalid metrics_client %q", clientType)
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
			return errors.New("datadog client environment variable 'DD_AGENT_HOST' must be set")
		}

		var err error
		activeClient, err = newDatadogClient(logger.With(zap.String("component", "datadog")), agent)
		if err != nil {
			return err
		}
	case Prometheus:
		activeClient = newPrometheusClient(logger.With(zap.String("component", "prometheus")))
	case NoOp:
	default:
		return errors.New("metrics client config not set")
	}
	return nil
}

func prefixMetricName(name string) string {
	return fmt.Sprintf("%s_%s", metricPrefix, name)
}

// Gauge measures the value of a metric at a particular time.
func Gauge(name string, value float64, tags ...Tag) {
	activeClient.Gauge(prefixMetricName(name), value, tags...)
}

// Count tracks how many times something happened.
func Count(name string, value int64, tags ...Tag) {
	activeClient.Count(prefixMetricName(name), value, tags...)
}

// Histogram tracks the statistical distribution of a set of values.
func Histogram(name string, value float64, tags ...Tag) {
	activeClient.Histogram(prefixMetricName(name), value, tags...)
}

// CreateHandler create an http.Handler if supported by the client.
// Otherwise an error will be returned.
func CreateHandler(log *zap.Logger) (http.Handler, error) {
	return activeClient.CreateHandler(log)
}

// Close the underlying client connection if supported.
func Close() error {
	return activeClient.Close()
}
